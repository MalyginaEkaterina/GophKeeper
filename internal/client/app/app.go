package app

import (
	"bufio"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MalyginaEkaterina/GophKeeper/internal/client"
	cpb "github.com/MalyginaEkaterina/GophKeeper/internal/client/proto"
	pb "github.com/MalyginaEkaterina/GophKeeper/internal/common/proto"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

var (
	errDataConflict      = errors.New("data conflict")
	errDataNotFound      = errors.New("data not found")
	errNeedLogin         = errors.New("auth data expired")
	errNeedFirstLogin    = errors.New("first login on server required")
	errAlreadyExists     = errors.New("already exists")
	errServerUnavailable = errors.New("server unavailable")
	errIncorrectAuth     = errors.New("incorrect auth data")
)

type App struct {
	stdin  *bufio.Reader
	log    *log.Logger
	config client.Config

	grpcClient *GrpcClient

	creds         Creds
	cache         *Cache
	offline       bool
	wasFirstLogin bool
}

func (a *App) prompt(message string) {
	username := a.creds.getUsername()

	if username != "" {
		if a.offline {
			fmt.Printf("(offline) %s@ %s > ", username, message)
		} else {
			fmt.Printf("%s@ %s > ", username, message)
		}
	} else {
		if a.offline {
			fmt.Printf("(offline) %s > ", message)
		} else {
			fmt.Printf("%s > ", message)
		}
	}
}

func (a *App) readString() string {
	s, err := a.stdin.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		panic("Failed to read from stdin")
	}
	return strings.TrimSpace(s)
}

// Start reads config from file or fills it in by default. Creates new Grpc Client, opens file for logs,
// then repeats doLogin or doKeep until it will be broken with exit command.
// At the end flushes cache into file.
func Start() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic("Getting user home directory error")
	}
	cfg := client.Config{
		ServerAddress:       "localhost:3200",
		GrpcTimeout:         30,
		FlushIntoFileAfter:  30,
		SyncWithServerAfter: 30,
		LogFilePath:         path.Join(dirname, "goph_keeper.log"),
		CacheFilePath:       path.Join(dirname, "goph_keeper.data"),
		PutsFilePath:        path.Join(dirname, "goph_keeper_put.data"),
		LoggedUserFilePath:  path.Join(dirname, "goph_keeper_user.data"),
	}

	configName := os.Getenv("CONFIG")
	if configName != "" {
		confData, err := os.ReadFile(configName)
		if err != nil {
			panic("Error while reading config file")
		}
		err = json.Unmarshal(confData, &cfg)
		if err != nil {
			panic("Error while parsing config file")
		}
	}

	app := App{
		stdin:  bufio.NewReader(os.Stdin),
		cache:  newCache(cfg),
		config: cfg,
	}

	app.grpcClient = newGrpcClient(cfg.ServerAddress, app.creds.getToken, time.Duration(cfg.GrpcTimeout)*time.Second)
	defer app.grpcClient.Close()

	flog, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer flog.Close()
	app.log = log.New(flog, "", log.LstdFlags)

	ctx, cancel := context.WithCancel(context.Background())
	for {
		app.doLogin(ctx)
		if !app.doKeep() {
			break
		}
	}
	cancel()
	app.flushCacheIntoFile()
}

func (a *App) flushCacheIntoFile() {
	password := a.creds.getPassword()
	if password != "" {
		if err := a.cache.flushIntoFile(password); err != nil {
			a.log.Printf("Flush cache into file error: %s\n", err.Error())
		}
	}
}

// runFlush runs in other goroutine after the first success login and flushes cache into file on ticks.
func (a *App) runFlush(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			a.flushCacheIntoFile()
		case <-ctx.Done():
			return
		}
	}
}

// runSyncWithServer runs in other goroutine after the first success login and fills cache from server on ticks.
func (a *App) runSyncWithServer(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			if a.creds.getToken() != "" {
				_ = a.fillCacheFromServer()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *App) startOffline() {
	fmt.Println("Entering offline mode")
	a.offline = true
	a.creds.clear()
}

// startOnline sends offline put requests to server if such requests exists.
func (a *App) startOnline() bool {
	fmt.Println("Entering online mode")
	a.offline = false
	if a.cache.hasPutRequests() {
		err := a.tryPutRequests()
		if errors.Is(err, errServerUnavailable) {
			fmt.Println("Server unavailable")
			a.startOffline()
			return false
		}
	}
	return true
}

func (a *App) doKeep() bool {
	for {
		if a.offline {
			a.prompt("Enter 'g' to get, 'p' to put, 'l' to get keys list, 'on' to return to online, 'q' to exit")
		} else {
			a.prompt("Enter 'g' to get, 'p' to put, 'l' to get keys list, 'q' to exit")
		}
		command := a.readString()
		if command != "q" {
			if a.creds.isExpired() {
				fmt.Println("Auth expired")
				a.flushCacheIntoFile()
				a.creds.clear()
				return true
			}
		}
		switch command {
		case "g":
			a.prompt("Enter key")
			key := a.readString()
			err := a.getByKey(key)
			if errors.Is(err, errNeedLogin) {
				return true
			} else if errors.Is(err, errServerUnavailable) {
				a.startOffline()
				return true
			}
		case "p":
			a.prompt("Enter key")
			key := a.readString()
			a.prompt("Enter metadata")
			md := a.readString()
			data := a.readData()
			err := a.putByKey(key, md, data)
			if errors.Is(err, errDataConflict) {
				fmt.Println("Conflicting updates")
				err = a.solveConflict(key, md, data)
				if errors.Is(err, errServerUnavailable) {
					a.startOffline()
					return true
				} else if errors.Is(err, errNeedLogin) {
					return true
				} else if err != nil {
					fmt.Println("Solving conflict error", err)
					continue
				}
			} else if errors.Is(err, errNeedLogin) {
				return true
			} else if errors.Is(err, errServerUnavailable) {
				a.startOffline()
				return true
			}
			fmt.Println("Put data completed")
		case "l":
			err := a.getKeyList()
			if errors.Is(err, errNeedLogin) {
				return true
			} else if errors.Is(err, errServerUnavailable) {
				a.startOffline()
				return true
			}
		case "on":
			if !a.offline {
				fmt.Println("Please choose 'g', 'p' or 'l'")
				continue
			}
			if a.doLoginOnServer() {
				if !a.startOnline() {
					return true
				}
			}
		case "q":
			return false
		default:
			fmt.Println("Please choose 'g', 'p' or 'l'")
		}
	}
}

func (a *App) putByKey(key string, md string, data []byte) error {
	version := a.cache.getNextVersionForKey(key)
	putReq := &pb.PutReq{
		Key:   key,
		Value: &pb.Value{Data: data, Metadata: md, Version: version},
	}
	if a.offline {
		a.cache.appendPutRequest(putReq)
	} else {
		err := a.grpcClient.putByKey(putReq)
		if errors.Is(err, errNeedLogin) {
			fmt.Println("Auth data expired")
			return err
		} else if errors.Is(err, errServerUnavailable) {
			fmt.Println("Server unavailable")
			return err
		} else if err != nil {
			fmt.Println("Got error from server while putting data")
			a.log.Printf("Got error from server while putting data: %s\n", err.Error())
			return err
		}
	}
	a.cache.put(key, putReq.Value)
	return nil
}

func (a *App) solveConflict(key string, newMetadata string, newData []byte) error {
	value, err := a.grpcClient.getByKey(key)
	if errors.Is(err, errServerUnavailable) {
		fmt.Println("Server unavailable")
		return err
	}
	if err != nil {
		return err
	}
	a.cache.put(key, value)
	fmt.Printf("Conflicting version on the server fok key %s:\n", key)
	a.printValue(value)
	fmt.Println("Offline version:")
	a.printValue(&pb.Value{Data: newData, Metadata: newMetadata})
	for {
		a.prompt("Enter 'u' to update, 'k' to keep the server version")
		command := a.readString()
		switch command {
		case "u":
			return a.putByKey(key, newMetadata, newData)
		case "k":
			return nil
		default:
			a.prompt("Please choose 'u' or 'k'")
		}
	}

}

func (a *App) fillCacheFromServer() error {
	lastVersion := a.cache.getVersion()
	serverData, currVersion, err := a.grpcClient.getAll(lastVersion)
	if errors.Is(err, errDataNotFound) {
		return nil
	} else if err != nil {
		a.log.Printf("Fill cache from server error: %s\n", err.Error())
		return err
	}
	a.cache.setCache(serverData, currVersion)
	return nil
}

func (a *App) getByKey(key string) error {
	var value *pb.Value
	var err error
	if a.offline {
		value = a.cache.getByKey(key)
	} else {
		value, err = a.grpcClient.getByKey(key)
	}
	if err != nil {
		if errors.Is(err, errNeedLogin) {
			fmt.Println("Auth data expired")
		} else if errors.Is(err, errDataNotFound) {
			fmt.Println("Data not found")
		} else if errors.Is(err, errServerUnavailable) {
			fmt.Println("Server unavailable")
		} else {
			a.log.Printf("Got error from server on getting by key: %s\n", err.Error())
		}
		return err
	}
	a.printValue(value)
	return nil
}

func (a *App) getKeyList() error {
	var keys []string
	var err error
	if a.offline {
		keys = a.cache.getKeys()
	} else {
		keys, err = a.grpcClient.getKeyList()
	}
	if err != nil {
		if errors.Is(err, errNeedLogin) {
			fmt.Println("Auth data expired")
		} else if errors.Is(err, errServerUnavailable) {
			fmt.Println("Server unavailable")
		} else {
			a.log.Printf("Got error from server on getting key list: %s\n", err.Error())
		}
		return err
	}
	fmt.Println("Keys:")
	for _, k := range keys {
		fmt.Println(k)
	}
	return nil
}

// auth does auth on server. At first login if there were offline put requests for logged user sends them to server.
// Then fills cache from server and flushes it into file.
func (a *App) auth(username, password string) error {
	token, err := a.grpcClient.auth(username, password)
	if errors.Is(err, errIncorrectAuth) {
		fmt.Println("Incorrect login/password")
		return err
	} else if errors.Is(err, errServerUnavailable) {
		a.startOffline()
		return err
	} else if err != nil {
		a.log.Printf("Got error from server on auth request: %s\n", err.Error())
		fmt.Println("Authentication failed")
		return err
	}
	a.creds.set(password, token)

	if !a.wasFirstLogin {
		a.creds.setUsername(username)
		usernameFromFile, err := a.cache.readUsernameFromFile()
		if err == nil {
			if usernameFromFile == username {
				err = a.cache.fillFromFile(password)
				if err != nil {
					fmt.Println("Local data decryption error:", err)
				}
				if a.cache.hasPutRequests() {
					err = a.tryPutRequests()
					if errors.Is(err, errServerUnavailable) {
						fmt.Println("Server unavailable")
						a.startOffline()
						return err
					}
				}
			}
		}
		err = a.cache.writeUsernameIntoFile(username)
		if err != nil {
			fmt.Println("Local data save error:", err)
		}
	}

	err = a.fillCacheFromServer()
	if errors.Is(err, errServerUnavailable) {
		a.startOffline()
		return err
	} else if err == nil {
		fmt.Println("Data synchronization completed")
	}
	err = a.cache.flushIntoFile(a.creds.getPassword())
	if err != nil {
		a.log.Printf("Flush cache into file error: %s\n", err.Error())
	}
	return nil
}

func (a *App) doLogin(ctx context.Context) {
	defer func() {
		if !a.wasFirstLogin {
			flushTicker := time.NewTicker(time.Duration(a.config.FlushIntoFileAfter) * time.Second)
			go a.runFlush(ctx, flushTicker)
			syncTicker := time.NewTicker(time.Duration(a.config.SyncWithServerAfter) * time.Second)
			go a.runSyncWithServer(ctx, syncTicker)
			a.wasFirstLogin = true
		}
	}()
	for {
		if !a.offline {
			if a.doLoginOnServer() {
				return
			}
		}
		if a.offline {
			if a.doOfflineLogin() {
				return
			}
		}
	}
}

func (a *App) readPassword() (string, error) {
	a.prompt("Enter password")
	passwordBin, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(passwordBin), nil
}

func (a *App) doOfflineLogin() bool {
	username, err := a.cache.readUsernameFromFile()
	if err != nil {
		fmt.Println("First login on server required")
		a.offline = false
		return false
	}

	password, err := a.readPassword()
	if err != nil {
		fmt.Println("Read password error:", err)
		return false
	}

	err = a.cache.fillFromFile(password)
	if errors.Is(err, errDecryption) {
		fmt.Println("Wrong password")
		return false
	} else if errors.Is(err, errNeedFirstLogin) {
		fmt.Println("First login on server required")
		a.offline = false
		return false
	} else if err != nil {
		fmt.Println("Local data decryption error")
		a.log.Printf("Fill cache from file error: %s\n", err.Error())
		return false
	}
	a.creds.setUsername(username)
	a.creds.set(password, "")
	return true
}

func (a *App) doLoginOnServer() bool {
	if !a.wasFirstLogin {
		a.prompt("Enter 'r' to register, 'l' to login")
		in := a.readString()
		switch in {
		case "r":
			a.prompt("Enter login")
			username := a.readString()
			password, err := a.readPassword()
			if err != nil {
				fmt.Println("Read password error:", err)
				return false
			}
			err = a.grpcClient.register(username, password)
			if errors.Is(err, errAlreadyExists) {
				fmt.Printf("User with login %s already exists", username)
				return false
			} else if err != nil {
				fmt.Println("Registration error", err)
				return false
			}
			err = a.auth(username, password)
			if err != nil {
				return false
			}
			return true
		case "l":
			a.prompt("Enter login")
			username := a.readString()
			password, err := a.readPassword()
			if err != nil {
				fmt.Println("Read password error:", err)
				return false
			}
			err = a.auth(username, password)
			if err != nil {
				return false
			}
			return true
		default:
			return false
		}
	} else {
		password, err := a.readPassword()
		if err != nil {
			fmt.Println("Read password error:", err)
			return false
		}
		err = a.auth(a.creds.getUsername(), password)
		if err != nil {
			return false
		}
		return true
	}
}

func (a *App) printValue(value *pb.Value) {
	if value == nil {
		fmt.Println("Data not found")
		return
	}
	var data cpb.Data
	err := proto.Unmarshal(value.Data, &data)
	if err != nil {
		fmt.Println("Parse data error", err)
		return
	}

	switch dt := data.Data.(type) {
	case *cpb.Data_PasswordData:
		fmt.Printf("login = %s, password = %s\nmetadata: %s\n",
			dt.PasswordData.Login, dt.PasswordData.Password, value.Metadata)
	case *cpb.Data_TextData:
		fmt.Printf("text = %s\nmetadata: %s\n", dt.TextData.Text, value.Metadata)
	case *cpb.Data_CardData:
		fmt.Printf("number = %s, expiry = %s, cvc = %s\nmetadata: %s\n",
			dt.CardData.Number, dt.CardData.Expiry, dt.CardData.Cvc, value.Metadata)
	case *cpb.Data_BinaryData:
		fmt.Printf("data = %s\nmetadata: %s\n", hex.EncodeToString(dt.BinaryData.Data), value.Metadata)
	}
}

func (a *App) readData() []byte {
	for {
		var data cpb.Data
		for {
			a.prompt("Enter type: 'p' for Password, 'c' for Card, 't' for Text, 'b' for binary data")
			switch a.readString() {
			case "p":
				a.prompt("Enter login")
				l := a.readString()
				a.prompt("Enter password")
				p := a.readString()
				data.Data = &cpb.Data_PasswordData{PasswordData: &cpb.Data_Password{Login: l, Password: p}}
			case "c":
				a.prompt("Enter number")
				n := a.readString()
				a.prompt("Enter expiry")
				exp := a.readString()
				a.prompt("Enter cvc")
				cvc := a.readString()
				data.Data = &cpb.Data_CardData{CardData: &cpb.Data_Card{Number: n, Expiry: exp, Cvc: cvc}}
			case "t":
				a.prompt("Enter text data")
				text := a.readString()
				data.Data = &cpb.Data_TextData{TextData: &cpb.Data_Text{Text: text}}
			case "b":
				a.prompt("Enter file name")
				binDataPath := a.readString()
				binData, err := os.ReadFile(binDataPath)
				if err != nil {
					fmt.Println("Read file error", err)
					continue
				}
				data.Data = &cpb.Data_BinaryData{BinaryData: &cpb.Data_Binary{Data: binData}}
			default:
				fmt.Println("Unknown type")
				continue
			}
			break
		}
		b, err := proto.Marshal(&data)
		if err != nil {
			fmt.Println("Create value error", err)
			continue
		}
		return b
	}
}

func (a *App) tryPutRequests() error {
	for {
		putReq, ok := a.cache.nextPutRequest()
		if !ok {
			break
		}
		err := a.grpcClient.putByKey(putReq)
		if errors.Is(err, errServerUnavailable) {
			return err
		} else if errors.Is(err, errDataConflict) {
			fmt.Println("Conflicting updates")
			err = a.solveConflict(putReq.Key, putReq.Value.Metadata, putReq.Value.Data)
			if err != nil {
				fmt.Println("Solving conflict error", err)
				a.log.Printf("Solving conflict error: %s\n", err.Error())
			}
		} else if err != nil {
			fmt.Println("Synchronization error", err)
			a.log.Printf("Got error from server while putting data: %s\n", err.Error())
		} else {
			a.cache.put(putReq.Key, putReq.Value)
		}
		a.cache.popPutRequest()
	}
	fmt.Println("Synchronization data with offline put requests completed")
	return nil
}
