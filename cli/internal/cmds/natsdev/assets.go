package natsdev

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	natsjwt "github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

type GenerateResult struct {
	AccountPublicKey  string
	AccountSeed       string
	SystemPublicKey   string
	OperatorPublicKey string
}

var Cmd = &cli.Command{
	Name:  "nats-dev-assets",
	Usage: "Generate ignored local NATS TLS/JWT dev files",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "Overwrite existing generated NATS dev files",
		},
	},
	Action: func(c *cli.Context) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		result, err := GenerateAssets(wd, c.Bool("force"))
		if err != nil {
			return err
		}
		pterm.Success.Println("Generated local NATS TLS/JWT assets")
		pterm.Info.Printfln("Account public key: %s", result.AccountPublicKey)
		pterm.Info.Println("Core can read the account seed from configs/nats/jwt/account.seed")
		pterm.Info.Println("Restart the nats container after regenerating these files")
		return nil
	},
}

func GenerateAssets(root string, force bool) (GenerateResult, error) {
	paths := assetPaths(root)
	required := []string{
		paths.caCert,
		paths.serverCert,
		paths.serverKey,
		paths.operatorJWT,
		paths.accountJWT,
		paths.accountSeed,
		paths.coreCreds,
		paths.systemCreds,
		paths.devEnv,
	}
	if !force {
		for _, path := range required {
			if _, err := os.Stat(path); err == nil {
				return GenerateResult{}, fmt.Errorf("%s already exists; use --force to regenerate dev NATS assets", path)
			} else if !os.IsNotExist(err) {
				return GenerateResult{}, err
			}
		}
	}

	if err := ensureDirs(paths); err != nil {
		return GenerateResult{}, err
	}
	if err := generateTLSAssets(paths); err != nil {
		return GenerateResult{}, err
	}
	result, err := generateJWTAssets(paths)
	if err != nil {
		return GenerateResult{}, err
	}
	return result, nil
}

type paths struct {
	certsDir     string
	credsDir     string
	jwtDir       string
	resolverDir  string
	caCert       string
	serverCert   string
	serverKey    string
	operatorJWT  string
	operatorSeed string
	accountJWT   string
	accountSeed  string
	systemJWT    string
	coreCreds    string
	systemCreds  string
	devEnv       string
}

func assetPaths(root string) paths {
	certsDir := filepath.Join(root, "configs", "nats", "certs")
	credsDir := filepath.Join(root, "configs", "nats", "creds")
	jwtDir := filepath.Join(root, "configs", "nats", "jwt")
	resolverDir := filepath.Join(root, "docker", "assets", "nats-jwt")
	return paths{
		certsDir:     certsDir,
		credsDir:     credsDir,
		jwtDir:       jwtDir,
		resolverDir:  resolverDir,
		caCert:       filepath.Join(certsDir, "ca.pem"),
		serverCert:   filepath.Join(certsDir, "server.pem"),
		serverKey:    filepath.Join(certsDir, "server-key.pem"),
		operatorJWT:  filepath.Join(jwtDir, "operator.jwt"),
		operatorSeed: filepath.Join(jwtDir, "operator.seed"),
		accountJWT:   filepath.Join(jwtDir, "account.jwt"),
		accountSeed:  filepath.Join(jwtDir, "account.seed"),
		systemJWT:    filepath.Join(jwtDir, "system-account.jwt"),
		coreCreds:    filepath.Join(credsDir, "core.creds"),
		systemCreds:  filepath.Join(credsDir, "sys.creds"),
		devEnv:       filepath.Join(root, "configs", "nats", "dev.env"),
	}
}

func ensureDirs(p paths) error {
	for _, dir := range []string{p.certsDir, p.credsDir, p.jwtDir, p.resolverDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}

func generateTLSAssets(p paths) error {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ca key: %w", err)
	}
	caTemplate := certificateTemplate("Cactus Local NATS CA", true)
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create ca cert: %w", err)
	}

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate server key: %w", err)
	}
	serverTemplate := certificateTemplate("localhost", false)
	serverTemplate.DNSNames = []string{"localhost", "nats"}
	serverTemplate.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create server cert: %w", err)
	}

	if err := writePEM(p.caCert, "CERTIFICATE", caDER, 0o644); err != nil {
		return err
	}
	if err := writePEM(p.serverCert, "CERTIFICATE", serverDER, 0o644); err != nil {
		return err
	}
	serverKeyDER, err := x509.MarshalECPrivateKey(serverKey)
	if err != nil {
		return fmt.Errorf("marshal server key: %w", err)
	}
	return writePEM(p.serverKey, "EC PRIVATE KEY", serverKeyDER, 0o600)
}

func certificateTemplate(commonName string, ca bool) *x509.Certificate {
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		serial = big.NewInt(time.Now().UnixNano())
	}
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"Cactus Local Dev"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().AddDate(10, 0, 0),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	if ca {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		template.BasicConstraintsValid = true
	} else {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}
	return template
}

func writePEM(path, blockType string, der []byte, perm os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	return pem.Encode(file, &pem.Block{Type: blockType, Bytes: der})
}

func generateJWTAssets(p paths) (GenerateResult, error) {
	operatorKP, err := nkeys.CreateOperator()
	if err != nil {
		return GenerateResult{}, fmt.Errorf("create operator key: %w", err)
	}
	defer operatorKP.Wipe()
	accountKP, err := nkeys.CreateAccount()
	if err != nil {
		return GenerateResult{}, fmt.Errorf("create account key: %w", err)
	}
	defer accountKP.Wipe()
	systemKP, err := nkeys.CreateAccount()
	if err != nil {
		return GenerateResult{}, fmt.Errorf("create system account key: %w", err)
	}
	defer systemKP.Wipe()

	operatorPublic, err := operatorKP.PublicKey()
	if err != nil {
		return GenerateResult{}, err
	}
	accountPublic, err := accountKP.PublicKey()
	if err != nil {
		return GenerateResult{}, err
	}
	accountSeed, err := accountKP.Seed()
	if err != nil {
		return GenerateResult{}, err
	}
	operatorSeed, err := operatorKP.Seed()
	if err != nil {
		return GenerateResult{}, err
	}
	systemPublic, err := systemKP.PublicKey()
	if err != nil {
		return GenerateResult{}, err
	}

	operatorClaims := natsjwt.NewOperatorClaims(operatorPublic)
	operatorClaims.Name = "cactus-local"
	operatorClaims.SystemAccount = systemPublic
	operatorJWT, err := operatorClaims.Encode(operatorKP)
	if err != nil {
		return GenerateResult{}, fmt.Errorf("encode operator jwt: %w", err)
	}

	accountJWT, err := accountClaimsJWT(accountPublic, "CACTUS", true, operatorKP)
	if err != nil {
		return GenerateResult{}, err
	}
	systemJWT, err := accountClaimsJWT(systemPublic, "SYS", false, operatorKP)
	if err != nil {
		return GenerateResult{}, err
	}
	coreCreds, err := userCreds("core", accountKP)
	if err != nil {
		return GenerateResult{}, err
	}
	systemCreds, err := userCreds("sys", systemKP)
	if err != nil {
		return GenerateResult{}, err
	}

	if err := os.WriteFile(p.operatorJWT, []byte(operatorJWT), 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(p.operatorSeed, operatorSeed, 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(p.accountJWT, []byte(accountJWT), 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(p.accountSeed, accountSeed, 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(p.systemJWT, []byte(systemJWT), 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(filepath.Join(p.resolverDir, accountPublic+".jwt"), []byte(accountJWT), 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(filepath.Join(p.resolverDir, systemPublic+".jwt"), []byte(systemJWT), 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(p.coreCreds, coreCreds, 0o600); err != nil {
		return GenerateResult{}, err
	}
	if err := os.WriteFile(p.systemCreds, systemCreds, 0o600); err != nil {
		return GenerateResult{}, err
	}
	devEnv := fmt.Sprintf("NATS_ACCOUNT_SEED=%s\nNATS_ACCOUNT_PUBLIC_KEY=%s\n", string(accountSeed), accountPublic)
	if err := os.WriteFile(p.devEnv, []byte(devEnv), 0o600); err != nil {
		return GenerateResult{}, err
	}

	return GenerateResult{
		AccountPublicKey:  accountPublic,
		AccountSeed:       string(accountSeed),
		SystemPublicKey:   systemPublic,
		OperatorPublicKey: operatorPublic,
	}, nil
}

func accountClaimsJWT(publicKey, name string, enableJetStream bool, signer nkeys.KeyPair) (string, error) {
	claims := natsjwt.NewAccountClaims(publicKey)
	claims.Name = name
	if enableJetStream {
		claims.Limits.JetStreamLimits.MemoryStorage = natsjwt.NoLimit
		claims.Limits.JetStreamLimits.DiskStorage = natsjwt.NoLimit
		claims.Limits.JetStreamLimits.Streams = natsjwt.NoLimit
		claims.Limits.JetStreamLimits.Consumer = natsjwt.NoLimit
	}
	return claims.Encode(signer)
}

func userCreds(name string, account nkeys.KeyPair) ([]byte, error) {
	userKP, err := nkeys.CreateUser()
	if err != nil {
		return nil, fmt.Errorf("create %s user key: %w", name, err)
	}
	defer userKP.Wipe()
	userPublic, err := userKP.PublicKey()
	if err != nil {
		return nil, err
	}
	userSeed, err := userKP.Seed()
	if err != nil {
		return nil, err
	}
	claims := natsjwt.NewUserClaims(userPublic)
	claims.Name = name
	userJWT, err := claims.Encode(account)
	if err != nil {
		return nil, fmt.Errorf("encode %s user jwt: %w", name, err)
	}
	creds, err := natsjwt.FormatUserConfig(userJWT, userSeed)
	if err != nil {
		return nil, fmt.Errorf("format %s creds: %w", name, err)
	}
	return creds, nil
}
