#!/usr/bin/env bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SDKMS_ENDPOINT="$ENDPOINT"
WORK_FOLDER="$(pwd)/$1"
ENV_FILE="$WORK_FOLDER.env"

mkdir -p "$WORK_FOLDER"

if [[ -f "$ENV_FILE" ]]; then
    echo "Read env from existed $ENV_FILE."
else
    echo "Env file does not exist, creating new one by performing test-setup."
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" test-setup > "$ENV_FILE"
    echo "Created new one at $ENV_FILE."
fi

source "$ENV_FILE"

connections_options=("10" "20" "50" "100" "200")
qps="9999"

pushd "$WORK_FOLDER"

for connections in "${connections_options[@]}"; do
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_KEY_ID" --mode CBC           | tee "${WORK_FOLDER}/c${connections}_enc_aes_256_cbc.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_KEY_ID" --mode CBC --decrypt | tee "${WORK_FOLDER}/c${connections}_dec_aes_256_cbc.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_KEY_ID" --mode FPE           | tee "${WORK_FOLDER}/c${connections}_enc_aes_256_fpe.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_KEY_ID" --mode FPE --decrypt | tee "${WORK_FOLDER}/c${connections}_dec_aes_256_fpe.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_KEY_ID" --mode GCM           | tee "${WORK_FOLDER}/c${connections}_enc_aes_256_gcm.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_KEY_ID" --mode GCM --decrypt | tee "${WORK_FOLDER}/c${connections}_dec_aes_256_gcm.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_192_KEY_ID" --mode FPE           | tee "${WORK_FOLDER}/c${connections}_enc_aes_192_fpe.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sym --kid "$TEST_AES_192_KEY_ID" --mode FPE --decrypt | tee "${WORK_FOLDER}/c${connections}_dec_aes_192_fpe.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load asym --kid "$TEST_RSA_KEY_ID"           | tee "${WORK_FOLDER}/c${connections}_enc_rsa_2048.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load asym --kid "$TEST_RSA_KEY_ID" --decrypt | tee "${WORK_FOLDER}/c${connections}_dec_rsa_2048.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load gen -t AES --size 256  | tee "${WORK_FOLDER}/c${connections}_gen_aes_256.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load gen -t RSA --size 2048 | tee "${WORK_FOLDER}/c${connections}_gen_rsa_2048.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load gen -t EC              | tee "${WORK_FOLDER}/c${connections}_gen_ec_nistp256.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load invoke --plugin-id "$TEST_HELLO_PLUGIN_ID"  | tee "${WORK_FOLDER}/c${connections}_plugin_hello.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sign --kid "$TEST_RSA_KEY_ID"                   | tee "${WORK_FOLDER}/c${connections}_sign_rsa_2048.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sign --kid "$TEST_RSA_KEY_ID" --verify          | tee "${WORK_FOLDER}/c${connections}_verify_rsa_2048.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sign --kid "$TEST_RSA_4096_KEY_ID"              | tee "${WORK_FOLDER}/c${connections}_sign_rsa_4096.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sign --kid "$TEST_RSA_4096_KEY_ID" --verify     | tee "${WORK_FOLDER}/c${connections}_verify_rsa_4096.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sign --kid "$TEST_EC_NIST_P256_KEY_ID"          | tee "${WORK_FOLDER}/c${connections}_sign_ec_nistp256.log"
    "$SCRIPT_DIR"/dsm-perf-tool --insecure -s "$SDKMS_ENDPOINT" -k "$TEST_API_KEY" -c "$connections" --qps "$qps" -d 120s load sign --kid "$TEST_EC_NIST_P256_KEY_ID" --verify | tee "${WORK_FOLDER}/c${connections}_verify_ec_nistp256.log"
done

popd
