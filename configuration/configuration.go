package configuration

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/azbuky/rosetta-vite/vite"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// Mode is the setting that determines if
// the implementation is "online" or "offline".
type Mode string

const (
	// Online is when the implementation is permitted
	// to make outbound connections.
	Online Mode = "ONLINE"

	// Offline is when the implementation is not permitted
	// to make outbound connections.
	Offline Mode = "OFFLINE"

	// Mainnet is the Vite Mainnet.
	Mainnet string = "MAINNET"

	// Testnet is the Vite Testnet.
	Testnet string = "TESTNET"

	// DataDirectory is the default location for all
	// persistent data.
	DataDirectory = "/data"

	// ModeEnv is the environment variable read
	// to determine mode.
	ModeEnv = "MODE"

	// NetworkEnv is the environment variable
	// read to determine network.
	NetworkEnv = "NETWORK"

	// PortEnv is the environment variable
	// read to determine the port for the Rosetta
	// implementation.
	PortEnv = "PORT"

	// GviteEnv is an optional environment variable
	// used to connect rosetta-vite to an already
	// running gvite node.
	GviteEnv = "GVITE"

	// InlineTransactions is an optional environmen variable
	// used to determine if transaction are returned inline
	// in /block or as other_transactions
	InlineTransactions = "INLINE_TXS"

	// DefaultGviteURL is the default URL for
	// a running gvite node. This is used
	// when GviteEnv is not populated.
	DefaultGviteURL = "http://localhost:48132/"

	// MiddlewareVersion is the version of rosetta-vite.
	MiddlewareVersion = "0.2.0"
)

// Configuration determines how
type Configuration struct {
	Mode    Mode
	Network *types.NetworkIdentifier
	//GenesisBlockIdentifier *types.BlockIdentifier
	GviteURL           string
	RemoteGvite        bool
	Port               int
	GviteArguments     string
	InlineTransactions bool
}

// LoadConfiguration attempts to create a new Configuration
// using the ENVs in the environment.
func LoadConfiguration() (*Configuration, error) {
	config := &Configuration{}

	modeValue := Mode(os.Getenv(ModeEnv))
	switch modeValue {
	case Online:
		config.Mode = Online
	case Offline:
		config.Mode = Offline
	case "":
		return nil, errors.New("MODE must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid mode", modeValue)
	}

	networkValue := os.Getenv(NetworkEnv)
	switch networkValue {
	case Mainnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: vite.Blockchain,
			Network:    vite.MainnetNetwork,
		}
		//config.GenesisBlockIdentifier = vite.MainnetGenesisBlockIdentifier
		config.GviteArguments = vite.MainnetGviteArguments
	case Testnet:
		config.Network = &types.NetworkIdentifier{
			Blockchain: vite.Blockchain,
			Network:    vite.TestnetNetwork,
		}
		//config.GenesisBlockIdentifier = vite.TestnetGenesisBlockIdentifier
		config.GviteArguments = vite.TestnetGviteArguments
	case "":
		return nil, errors.New("NETWORK must be populated")
	default:
		return nil, fmt.Errorf("%s is not a valid network", networkValue)
	}

	config.GviteURL = DefaultGviteURL
	envGviteURL := os.Getenv(GviteEnv)
	if len(envGviteURL) > 0 {
		config.RemoteGvite = true
		config.GviteURL = envGviteURL
	}

	portValue := os.Getenv(PortEnv)
	if len(portValue) == 0 {
		return nil, errors.New("PORT must be populated")
	}

	config.InlineTransactions = vite.InlineTransactions
	inlineTransactions := os.Getenv(InlineTransactions)
	if len(inlineTransactions) > 0 {
		inline, err := strconv.ParseBool(inlineTransactions)
		if err == nil {
			config.InlineTransactions = inline
		}
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || len(portValue) == 0 || port <= 0 {
		return nil, fmt.Errorf("%w: unable to parse port %s", err, portValue)
	}
	config.Port = port

	return config, nil
}
