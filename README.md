# 🛡 Go Secure Folder Backup To S3

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="GitHub license" /></a>
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.17+-00ADD8.svg" alt="Go Version" /></a>
</p>

> A Go application that zips, encrypts, and securely backup folder to an AWS S3 Bucket.

---

## 🌟 Features

- **🔒 Secure**: Uses trusted public keys for encryption and server's private keys for signing.
- **🗂 Batch Processing**: Zips an entire folder for encrypted transfer.
- **☁️ Cloud Ready**: Easily uploads to AWS S3 Bucket.

---

## 📦 Prerequisites

- [Go](https://golang.org/) (version 1.21 or higher)
- [AWS CLI](https://aws.amazon.com/cli/) (configured with required access permissions)
- [PGP keys](https://www.openpgp.org/) (a trusted public key for encryption and a private key for server authentication)

---

## 🚀 Quick Start

### Clone the repository

```bash
git clone https://github.com/thomasdseao/go-secure-backup-s3.git
```

### Navigate and build

```bash
cd go-secure-backup-s3
go build .
```

### Execute

```bash
./go-secure-backup-s3 -folder="/tmp/testpath" -tpubkey="/tmp/public_key.asc" -sprivkey="/tmp/private_key.asc" -bucket="backup-timestamp" -s3key="testbackup" -aws-access-key="XXXXXXXXXXXXXXXXX" -aws-secret-key="XXXXXXXXXXXXXX" -aws-region="eu-north-1" -signingpassword="XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

### CLI Flags

- `folder`: Path to the folder you want to zip.
- `tpubkey`: Path to the trusted public key for encryption.
- `sprivkey`: Path to the server's private key for signing.
- `signingpassword`: Password for the server's private key.
- `bucket`: Name of the S3 bucket where the file will be uploaded.
- `s3key`: Key name to use in S3.
- `aws-access-key`: AWS access key for S3 authentication.
- `aws-secret-key`: AWS secret key for S3 authentication.


---

## 📃 License

This project is licensed under the MIT License.

---

## 🙋‍♂️ Questions

For questions and support, please create an issue.