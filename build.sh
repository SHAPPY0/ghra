# GOOS=linux GOARCH=amd64 go build -o ./build/ghra_linux_v1.0
# GOOS=windows GOARCH=amd64 go build -o ./build/ghra_win_v1.0.exe
# GOOS=darwin GOARCH=amd64 go build -o ./build/ghra_mac_v1.0

VERSION=$1
OUTPUT_DIR="build"

if [ ! "$VERSION" ]; then
    echo -ne "\nMissing Version.\n"
    exit 1
fi

TARGETS=(
  "linux/amd64:tar.gz"
  "linux/arm64:tar.gz"
  "darwin/amd64:tar.gz"
  "windows/amd64:zip"
)

for target in "${TARGETS[@]}" do
    goos_arch="${target%:*}"
    file_extension="${target#*:}"
    IFS='/' read -r goos goarch <<< "$goos_arch"

    # Set the environment variables
    export GOOS="$goos"
    export GOARCH="$goarch"

    # Build the package
    echo "Building for $GOOS/$GOARCH..."
    file_name="${PACKAGE_NAME}-${PACKAGE_VERSION}-${GOOS}-${GOARCH}"
    output_file="$OUTPUT_DIR/$file_name"

    if [ $GOOS = "windows" ]; then
            output_zip="${output_file}"
            output_file+='.exe'
            GOOS=$GOOS GOARCH=$GOARCH go build -o "${output_file}"
            zip -m "${output_zip}.zip" "${output_file}"
        else
        GOOS=$GOOS GOARCH=$GOARCH go build -o "${output_file}"
        tar -czvf "${output_file}.tar.gz" "${output_file}"
    fi
done
