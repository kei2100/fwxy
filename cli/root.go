package cli

import (
	"net/http"
	"strings"

	"log"

	"errors"
	"net"
	"net/url"

	"github.com/kei2100/h-fwd/config"
	"github.com/kei2100/h-fwd/hfwd"
	"github.com/spf13/cobra"
)

var (
	// option parameters for the url configuration
	rewritePaths []string
)

var (
	// option parameters for the headers configuration
	headers  []string
	username string
	password string
)

var (
	// options parameters for the client certification
	caCertPath     string
	pkcs12Path     string
	pkcs12Password string
)

var (
	// TODO flags
	lnAddr = "127.0.0.1"
	lnPort = "8080"
)

func init() {
	flags := RootCmd.PersistentFlags()

	flags.StringVarP(&username, "username", "u", "", "username for the basic authentication")
	flags.StringVarP(&password, "password", "p", "", "password for the basic authentication")
	flags.StringSliceVarP(&rewritePaths, "rewrite", "r", []string{}, "list for path rewrite (-r /old:/new -r /o:/n OR -r /old:/new,/o:/n)")

	flags.StringSliceVarP(&headers, "header", "H", []string{}, "list for the additional http headers (-H Host:https://custom.example.com -H 'User-Agent:My Agent'")

	flags.StringVar(&caCertPath, "ca-cert", "", "path of the additional CA certificate PEM")
	flags.StringVar(&pkcs12Path, "pkcs12", "", "path of the PKCS12 encoded file for the client certification")
	flags.StringVar(&pkcs12Password, "pkcs12-password", "", "password for the PKCS12 file")
}

// RootCmd for CLI
var RootCmd = &cobra.Command{
	Use:   "hfwd <destination URL>",
	Short: "hfwd is a simple HTTP forward proxy",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("requires at the <destination URL>")
		}
		return nil
	},
	//Example: "", // TODO
	Run: func(cmd *cobra.Command, args []string) {
		dst, err := url.Parse(args[0])
		if err != nil {
			log.Fatalf("failed to parse the <desitination URL>: %v", err)
		}

		params := config.Parameters{}
		params.RewritePaths = parseRewritePaths(rewritePaths)

		params.Header = parseHeaders(headers)
		params.Username = username
		params.Password = password

		params.CACertPath = caCertPath
		params.PKCS12Path = pkcs12Path
		params.PKCS12Password = pkcs12Password

		if err := params.Load(); err != nil {
			log.Fatalf("failed to load configuration: %v", err)
		}

		handler, err := hfwd.NewHandler(dst, &params)
		if err != nil {
			log.Fatalf("failed to setup the foward proxy: %v", err)
		}

		ln, err := net.Listen("tcp", lnAddr+":"+lnPort)
		if err != nil {
			log.Fatalf("failed to listening start at %v:%v: %v", lnAddr, lnPort, err)
		}
		defer ln.Close()

		log.Printf("hfwd listening on %v:%v", lnAddr, lnPort)
		out := http.Serve(ln, handler)

		log.Println(out)
	},
}

func parseRewritePaths(rewritePaths []string) map[string]string {
	m := make(map[string]string, len(rewritePaths))
	for _, p := range rewritePaths {
		sp := strings.SplitAfterN(p, ":", 2)
		m[sp[0]] = sp[1]
	}
	return m
}

func parseHeaders(headers []string) http.Header {
	hh := make(http.Header, len(headers))
	for _, h := range headers {
		sp := strings.SplitAfterN(h, ":", 2)
		hh.Add(sp[0], sp[1])
	}
	return hh
}
