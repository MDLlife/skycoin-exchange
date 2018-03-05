package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"os"

	"net/http"

	logging "github.com/op/go-logging"
	"github.com/skycoin/skycoin-exchange/src/coin/aynrandcoin"
	"github.com/skycoin/skycoin-exchange/src/coin/bitcoin"
	"github.com/skycoin/skycoin-exchange/src/coin/fishercoin"
	"github.com/skycoin/skycoin-exchange/src/coin/lifecoin"
	"github.com/skycoin/skycoin-exchange/src/coin/metalicoin"
	"github.com/skycoin/skycoin-exchange/src/coin/mzcoin"
	"github.com/skycoin/skycoin-exchange/src/coin/shellcoin"
	"github.com/skycoin/skycoin-exchange/src/coin/skycoin"
	"github.com/skycoin/skycoin-exchange/src/coin/suncoin"
	"github.com/skycoin/skycoin-exchange/src/server"
	"github.com/skycoin/skycoin/src/cipher"
)

var (
	secKey     = "38d010a84c7b9374352468b41b076fa585d7dfac67ac34adabe2bbba4f4f6257"
	logger     = logging.MustGetLogger("exchange.main")
	logFormat  = "[%{module}:%{level}] %{message}"
	logModules = []string{
		"exchange.main",
		"exchange.server",
		"exchange.account",
		"exchange.api",
		"exchange.bitcoin",
		"exchange.skycoin",
		"exchange.gin",
	}
)

func registerFlags(cfg *server.Config) {
	flag.StringVar(&cfg.Server, "server", "127.0.0.1", "server ip")
	flag.IntVar(&cfg.Port, "port", 8080, "server listen port")
	flag.IntVar(&cfg.BtcFee, "btc-fee", 10000, "transaction fee in satoish")
	flag.StringVar(&cfg.DataDir, "data-dir", ".skycoin-exchange", "data directory")
	flag.StringVar(&cfg.Seed, "seed", "", "wallet's seed")
	flag.IntVar(&cfg.UtxoPoolSize, "poolsize", 1000, "utxo pool size")
	flag.StringVar(&cfg.Admins, "admins", "", "admin pubkey list")
	var (
		skyNodeAddr        string
		mzNodeAddr         string
		shellNodeAddr      string
		sunNodeAddr        string
		aynNodeAddr        string
		metaliNodeAddr     string
		lifecoinNodeAddr   string
		fishercoinNodeAddr string
	)
	flag.StringVar(&skyNodeAddr, "skycoin-node-addr", "127.0.0.1:6420", "skycoin node address")
	flag.StringVar(&mzNodeAddr, "mzcoin-node-addr", "127.0.0.1:7420", "mzcoin node address")
	flag.StringVar(&shellNodeAddr, "shellcoin-node-addr", "127.0.0.1:7520", "suncoin node address")
	flag.StringVar(&sunNodeAddr, "suncoin-node-addr", "127.0.0.1:7620", "suncoin node address")
	flag.StringVar(&aynNodeAddr, "aynrandcoin-node-addr", "127.0.0.1:7720", "suncoin node address")
	flag.StringVar(&metaliNodeAddr, "metalicoin-node-addr", "127.0.0.1:7820", "metalicoin node address")
	flag.StringVar(&lifecoinNodeAddr, "lifecoin-node-addr", "127.0.0.1:8420", "lifecoin node address")
	flag.StringVar(&fishercoinNodeAddr, "fishercoin-node-addr", "127.0.0.1:8520", "fishercoin node address")
	flag.BoolVar(&cfg.HTTPProf, "http-prof", false, "enable http profiling")
	flag.StringVar(&cfg.Seckey, "seckey", "38d010a84c7b9374352468b41b076fa585d7dfac67ac34adabe2bbba4f4f6257", "private key used for encrypting and decryping messages")

	flag.Set("logtostderr", "true")
	flag.Parse()
	cfg.NodeAddresses[skycoin.Type] = skyNodeAddr
	cfg.NodeAddresses[mzcoin.Type] = mzNodeAddr
	cfg.NodeAddresses[shellcoin.Type] = shellNodeAddr
	cfg.NodeAddresses[suncoin.Type] = sunNodeAddr
	cfg.NodeAddresses[aynrandcoin.Type] = aynNodeAddr
	cfg.NodeAddresses[metalicoin.Type] = metaliNodeAddr
	cfg.NodeAddresses[lifecoin.Type] = lifecoinNodeAddr
	cfg.NodeAddresses[fishercoin.Type] = fishercoinNodeAddr
}

func main() {
	initLogging(logging.DEBUG, true)
	cfg := initConfig()
	initProfiling(cfg.HTTPProf)

	// print pubkey so that client can use that to communicate with server
	sk := cipher.MustSecKeyFromHex(cfg.Seckey)
	logger.Info("pubkey:%v", cipher.PubKeyFromSecKey(sk).Hex())

	s := server.New(cfg)
	// Bind supported coins
	s.BindCoins(
		&bitcoin.Bitcoin{},
		skycoin.New(cfg.NodeAddresses[skycoin.Type]),
		mzcoin.New(cfg.NodeAddresses[mzcoin.Type]),
		shellcoin.New(cfg.NodeAddresses[shellcoin.Type]),
		aynrandcoin.New(cfg.NodeAddresses[aynrandcoin.Type]),
		suncoin.New(cfg.NodeAddresses[suncoin.Type]),
		lifecoin.New(cfg.NodeAddresses[lifecoin.Type]),
		fishercoin.New(cfg.NodeAddresses[fishercoin.Type]),
		metalicoin.New(cfg.NodeAddresses[metalicoin.Type]))
	s.Run()
}

func initConfig() *server.Config {
	cfg := server.NewConfig()
	registerFlags(cfg)
	if cfg.Seed == "" {
		flag.Usage()
		panic("seed must be set")
	}

	return cfg
}

func initLogging(level logging.Level, color bool) {
	format := logging.MustStringFormatter(logFormat)
	logging.SetFormatter(format)
	bk := logging.NewLogBackend(os.Stdout, "", 0)
	bk.Color = true
	bkLvd := logging.AddModuleLevel(bk)
	for _, s := range logModules {
		bkLvd.SetLevel(level, s)
	}

	logging.SetBackend(bkLvd)
}

func initProfiling(httpProf bool) {
	if httpProf {
		go func() {
			log.Println(http.ListenAndServe(":6061", nil))
		}()
	}
}
