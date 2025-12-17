# #!/bin/bash
# set -e

# COVER_FILE="coverage.out"

# echo "🔹 Generating list of packages for coverage (exclude entities & servers)..."
# coverpkgs=$(go list ./modules/... | grep -vE 'modules/(entities|servers)' | tr '\n' ',' | sed 's/,$//')

# echo "🔹 Generating list of packages to test (exclude entities & servers, only with *_test.go)..."
# testpkgs=""
# for pkg in $(go list ./modules/... | grep -vE 'modules/(entities|servers)'); do
#   fs_path=$(echo "$pkg" | sed 's|^p2p-back-end/||') # แปลงเป็น relative path ของ filesystem
#   if ls "$fs_path"/*_test.go >/dev/null 2>&1; then
#     testpkgs="$testpkgs $pkg"
#   fi
# done

# echo "🔹 Running tests with coverage..."
# go test -v -coverpkg=$coverpkgs -coverprofile=$COVER_FILE $testpkgs

# echo "🔹 Opening coverage report..."
# go tool cover -html=$COVER_FILE -o coverage.html
# echo "✅ Report saved to coverage.html"
#!/bin/bash
set -e

# 1. กำหนดตัวแปรโฟลเดอร์และชื่อไฟล์
OUT_DIR="testResult"
COVER_FILE="$OUT_DIR/coverage.out"
HTML_FILE="$OUT_DIR/coverage.html"

# 2. สร้างโฟลเดอร์ testResult ถ้ายังไม่มี (mkdir -p จะไม่ error ถ้ามีอยู่แล้ว)
echo "🔹 Creating output directory: $OUT_DIR..."
mkdir -p $OUT_DIR

echo "🔹 Generating list of packages for coverage (exclude entities & servers)..."
coverpkgs=$(go list ./modules/... | grep -vE 'modules/(entities|servers)' | tr '\n' ',' | sed 's/,$//')

echo "🔹 Generating list of packages to test (exclude entities & servers, only with *_test.go)..."
testpkgs=""
for pkg in $(go list ./modules/... | grep -vE 'modules/(entities|servers)'); do
  fs_path=$(echo "$pkg" | sed 's|^p2p-back-end/||') 
  if ls "$fs_path"/*_test.go >/dev/null 2>&1; then
    testpkgs="$testpkgs $pkg"
  fi
done

echo "🔹 Running tests with coverage..."
# 3. รันเทสและเก็บผลลัพธ์ลงใน testResult/coverage.out
go test -v -coverpkg=$coverpkgs -coverprofile=$COVER_FILE $testpkgs

echo "🔹 Generating HTML report..."
# 4. สร้างไฟล์ HTML ไว้ใน testResult/coverage.html
go tool cover -html=$COVER_FILE -o $HTML_FILE

echo "✅ Success! Report saved at: $HTML_FILE"