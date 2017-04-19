# cd into this repo first
# chmod 550 build.sh
FRAMEWORK_NAME="RTCSignaling"
DERIVED_DATA_DIR="./build/derivedData"
#UNIVERSAL_BUILD_DIR="./build/universal"
DEVICE_BUILD_DIR="./build/device"
#SIMULATOR_BUILD_DIR="./build/simulator"

# Clean up previous build
if [ -d "./build" ]; then
rm -rf "./build"
fi

# OS Build
xcodebuild -project "${FRAMEWORK_NAME}/${FRAMEWORK_NAME}.xcodeproj" -scheme ${FRAMEWORK_NAME} -configuration Release -arch arm64 -arch armv7 -arch armv7s ONLY_ACTIVE_ARCH=NO -sdk "iphoneos" -derivedDataPath "${DERIVED_DATA_DIR}" clean build

mkdir -p "${DEVICE_BUILD_DIR}"

# Copy framework files
cp -r "${DERIVED_DATA_DIR}/Build/Products/Release-iphoneos/${FRAMEWORK_NAME}.framework" "${DEVICE_BUILD_DIR}/${FRAMEWORK_NAME}.framework"

# Lipo up
# lipo -create -output "${UNIVERSAL_BUILD_DIR}/${FRAMEWORK_NAME}.framework/${FRAMEWORK_NAME}" "${DEVICE_BUILD_DIR}/${FRAMEWORK_NAME}.framework/${FRAMEWORK_NAME}" "${SIMULATOR_BUILD_DIR}/${FRAMEWORK_NAME}.framework/${FRAMEWORK_NAME}" 

rm -rf "${DERIVED_DATA_DIR}"