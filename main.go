package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	var (
		folderPath, trustedPublicKeyPath, serverPrivateKeyPath, signingKeyPassword,
		bucketName, s3KeyName, awsRegion, awsAccessKey, awsSecretKey string
	)

	flag.StringVar(&folderPath, "folder", "", "Path to the folder to zip")
	flag.StringVar(&trustedPublicKeyPath, "tpubkey", "", "Path to the trusted public key for encryption")
	flag.StringVar(&serverPrivateKeyPath, "sprivkey", "", "Path to the server's private key for signing")
	flag.StringVar(&signingKeyPassword, "signingpassword", "", "Password for the signing key")
	flag.StringVar(&bucketName, "bucket", "", "S3 bucket name")
	flag.StringVar(&s3KeyName, "s3key", "", "Key name to use in S3")
	flag.StringVar(&awsRegion, "aws-region", "", "AWS bucket region")
	flag.StringVar(&awsAccessKey, "aws-access-key", "", "AWS access key")
	flag.StringVar(&awsSecretKey, "aws-secret-key", "", "AWS secret key")
	flag.Parse()

	// Validate input parameters
	if err := validateInputs(folderPath, trustedPublicKeyPath, serverPrivateKeyPath, signingKeyPassword, bucketName, s3KeyName, awsRegion, awsAccessKey, awsSecretKey); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Zip the folder
	zipBuffer := new(bytes.Buffer)
	if err := ZipFolder(folderPath, zipBuffer); err != nil {
		panic(err)
	}

	// Encrypt and sign the data
	encryptedBuffer := new(bytes.Buffer)
	if err := EncryptAndSign(zipBuffer, trustedPublicKeyPath, serverPrivateKeyPath, encryptedBuffer, signingKeyPassword); err != nil {
		panic(err)
	}

	// Upload to S3
	if err := UploadToS3(encryptedBuffer, bucketName, s3KeyName, awsRegion, awsAccessKey, awsSecretKey); err != nil {
		panic(err)
	}

	fmt.Printf("Successfully uploaded data to %s/%s\n", bucketName, s3KeyName)
}

// validateInputs performs validation on input parameters and returns an error if validation fails.
// Parameters:
// - folderPath: Path to the folder to zip
// - trustedPublicKeyPath: Path to the trusted public key for encryption
// - serverPrivateKeyPath: Path to the server's private key for signing
// - signingKeyPassword: Password for the signing key
// - bucketName: S3 bucket name
// - s3KeyName: Key name to use in S3
// - awsRegion: AWS bucket region
// - awsAccessKey: AWS access key
// - awsSecretKey: AWS secret key
func validateInputs(folderPath, trustedPublicKeyPath, serverPrivateKeyPath, signingKeyPassword, bucketName, s3KeyName, awsRegion, awsAccessKey, awsSecretKey string) error {
	// Validation logic for input parameters
	if folderPath == "" || !isDirExists(folderPath) {
		return fmt.Errorf("Invalid or missing folder path")
	}

	if trustedPublicKeyPath == "" || !isFileExists(trustedPublicKeyPath) {
		return fmt.Errorf("Invalid or missing trusted public key path")
	}

	if serverPrivateKeyPath == "" || !isFileExists(serverPrivateKeyPath) {
		return fmt.Errorf("Invalid or missing server private key path")
	}

	if signingKeyPassword == "" {
		return fmt.Errorf("Invalid or missing signing key password")
	}

	if awsRegion == "" {
		return fmt.Errorf("Invalid or missing S3 bucket region")
	}

	if bucketName == "" || s3KeyName == "" {
		return fmt.Errorf("Invalid or missing S3 bucket name or key name")
	}

	if awsAccessKey == "" || awsSecretKey == "" {
		return fmt.Errorf("Invalid or missing AWS access key or secret key")
	}

	return nil
}

// isDirExists checks if a directory exists at the specified path.
func isDirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// isFileExists checks if a file exists at the specified path.
func isFileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// ZipFolder creates a ZIP archive of the given folder.
// Parameters:
// - folderPath: Path to the folder to be zipped
// - zipBuffer: Buffer to write the ZIP archive to
func ZipFolder(folderPath string, zipBuffer *bytes.Buffer) error {
	// Create a new zip archive.
	zipWriter := zip.NewWriter(zipBuffer)

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			f, err := zipWriter.Create(strings.TrimPrefix(path, folderPath+"/"))
			if err != nil {
				return err
			}

			_, err = f.Write(data)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Close the archive.
	return zipWriter.Close()
}

// EncryptAndSign encrypts the zipBuffer using the trustedPublicKeyPath and signs it using the serverPrivateKeyPath.
// Parameters:
// - zipBuffer: The data to be encrypted and signed
// - trustedPublicKeyPath: Path to the trusted public key for encryption
// - serverPrivateKeyPath: Path to the server's private key for signing
// - encryptedBuffer: Buffer to write the encrypted data to
// - signingKeyPassword: Password for decrypting the signing private key
func EncryptAndSign(zipBuffer *bytes.Buffer, trustedPublicKeyPath string, serverPrivateKeyPath string, encryptedBuffer *bytes.Buffer, signingKeyPassword string) error {
	// Read the trusted public key
	publicKeyFile, err := os.Open(trustedPublicKeyPath)
	if err != nil {
		return err
	}
	defer publicKeyFile.Close()

	publicKeyList, err := openpgp.ReadArmoredKeyRing(publicKeyFile)
	if err != nil {
		return err
	}

	// Read the server private key
	privateKeyFile, err := os.Open(serverPrivateKeyPath)
	if err != nil {
		return err
	}
	defer privateKeyFile.Close()

	privateKeyList, err := openpgp.ReadArmoredKeyRing(privateKeyFile)
	if err != nil {
		return err
	}

	// Decrypt the server private key using the provided password
	for _, entity := range privateKeyList {
		if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
			// Check if the private key is encrypted
			err := entity.PrivateKey.Decrypt([]byte(signingKeyPassword))
			if err != nil {
				return err
			}
		}
	}

	// Encrypt and Sign
	w, err := openpgp.Encrypt(encryptedBuffer, publicKeyList, privateKeyList[0], nil, nil)
	if err != nil {
		return err
	}

	_, err = w.Write(zipBuffer.Bytes())
	if err != nil {
		return err
	}

	// Close the writer to finalize encryption.
	return w.Close()
}

// UploadToS3 uploads the encryptedBuffer to S3.
// Parameters:
// - encryptedBuffer: The encrypted data to be uploaded
// - bucketName: S3 bucket name
// - fileName: Key name to use in S3
// - region: AWS bucket region
// - accessKey: AWS access key
// - secretKey: AWS secret key
func UploadToS3(encryptedBuffer *bytes.Buffer, bucketName, fileName, region, accessKey, secretKey string) error {
	// Create AWS credentials using access key and secret key
	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")

	// Set up AWS session with your credentials
	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(region), // Replace with your desired AWS region
		Credentials: creds,
	})

	if err != nil {
		panic(err)
	}

	svc := s3.New(awsSession)

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(encryptedBuffer.Bytes()),
	})
	return err
}