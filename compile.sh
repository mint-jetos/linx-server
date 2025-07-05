# 1️⃣ Clone the repository
git clone https://github.com/andreimarcu/linx-server.git
cd linx-server

# 2️⃣ Initialize Go modules
go mod init github.com/andreimarcu/linx-server
go mod tidy

# 3️⃣ Install rice tool (used for embedding assets)
go install github.com/GeertJohan/go.rice/rice@latest

# 4️⃣ Add rice to your shell's PATH
export PATH=$PATH:$HOME/go/bin   # Or add this to ~/.bashrc to make it permanent

# 5️⃣ Embed static assets (from root of the project)
rice embed-go

# 6️⃣ Build the binary
go build -o linx-server

# 7️⃣ Append embedded assets to the binary (critical step!)
rice append --exec linx-server

# 8️⃣ Run the server with your config file
cp linx-server.conf.example linx-server.conf
./linx-server -config linx-server.conf

#go clean -modcache
#rm -rf ~/go
