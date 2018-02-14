package serverconf

import (
	"log"
)

var murmurCompatRules = map[string]string{
	"logfile":            "LogPath",
	"welcometext":        "WelcomeText",
	"port":               "Port",
	"host":               "Address",
	"serverpassword":     "ServerPassword",
	"bandwidth":          "MaxBandwidth",
	"users":              "MaxUsers",
	"textmessagelength":  "MaxTextMessageLength",
	"imagemessagelength": "MaxImageMessageLength",
	"allowhtml":          "AllowHTML",
	"sslCert":            "CertPath",
	"sslKey":             "KeyPath",
	"sendversion":        "SendOSInfo",
	"allowping":          "AllowPing",
	"usersperchannel":    "MaxUsersPerChannel",
	"defaultchannel":     "DefaultChannel",
	"rememberchannel":    "RememberChannel",
	"registerName":       "RegisterName",
	"registerHostname":   "RegisterHost",
	"registerPassword":   "RegisterPassword",
	"registerUrl":        "RegisterWebUrl",
	"registerLocation":   "RegisterLocation",
}

// TranslateMurmur converts a source map with supported options from a murmur.ini
// into a normal config map. It also emits warnings for some common, but unsupported
// Murmur options.
func TranslateMurmur(source map[string]string) (target map[string]string) {
	log.Println("Using Murmur compatibility mode for configuration file")
	target = make(map[string]string)
	for kmurmur, v := range source {
		if kgrumble, ok := murmurCompatRules[kmurmur]; ok {
			target[kgrumble] = v
		}
		switch kmurmur {
		case "database":
			log.Println("* Grumble does not yet support Murmur databases directly (see issue #21 on github).")
			if driver, ok := source["dbDriver"]; !ok || driver == "QSQLITE" {
				log.Println("  To convert a previous SQLite database, use the --import-murmurdb flag.")
			}
		case "sslDHParams":
			log.Println("* Go does not implement DHE modes in TLS, so the configured dhparams are ignored.")
		case "sslCiphers":
			log.Println("* Support for changing TLS ciphers is not implemented yet.")
		case "ice":
			log.Println("* Grumble does not support ZeroC ICE.")
		case "grpc":
			log.Println("* Grumble does not yet support gRPC (see issue #23 on github).")
		}
	}
	return target
}
