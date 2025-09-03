#!/bin/bash

modules_code_dir="$PWD/modules"
modules_compiled_dir="$PWD/tools/compiled-modules"
linuxmodules=(
  "cat"
  "download"
  "ls"
  "shell"
  "run"
)

windowsmodules=(
  "cat"
  "download"
  "ls"
  "shell"
  "run"
  "psexec"
  "powercliff"
  "evasion"
  "execute-assembly"
  "procdump"
  "token"
  "inline-clr"
  "runas"
  "ps"
)

#sudo apt install protoc-gen-go-grpc
#sudo apt install protoc-gen-go
#protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protobuf/orstedrpc/orstedrpc.proto
#protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protobuf/eventrpc/eventrpc.proto

compile_protobuf() {
    echo
    echo "########################"
    echo "Compiling Protobuf Files"
    echo "########################"
    echo

    proto_files=(
        "protobuf/orstedrpc/orstedrpc.proto"
        "protobuf/eventrpc/eventrpc.proto"
    )

    for proto in "${proto_files[@]}"; do
        echo "[+] Compiling $proto"
        if ! protoc \
            --go_out=. --go_opt=paths=source_relative \
            --go-grpc_out=. --go-grpc_opt=paths=source_relative \
            "$proto"; then
            echo "[-] Failed to compile $proto"
            return 1
        fi
        echo "[+] Done $proto"
    done

    echo
    echo "###########################"
    echo "All protobuf files compiled "
    echo "###########################"
    echo
}


compile_client_server() {
    echo
    echo "##################################"
    echo "Compiling Orsted Server and Client"
    echo "##################################"
    echo

    echo "[+] Compiling server ..."
    if ! go build -ldflags="-s -w" -o orsted-client client/main.go; then
        echo "[-] Failed to compile server"
        return 1
    fi
    echo "[+] Server compiled successfully"

    echo "[+] Compiling client ..."
    if ! go build -ldflags="-s -w" -o orsted-server server/main.go; then
        echo "[-] Failed to compile client"
        return 1
    fi
    echo "[+] Client compiled successfully"

    echo
    echo "##############"
    echo "Done compiling"
    echo "##############"
    echo
}


compile_linux_module() {
    module_name=$1
    module_dir="$modules_code_dir/$module_name"
    output_dir="$modules_compiled_dir/linux/$module_name.so"
    echo "[+] Compiling Linux Module $module_name"

    if cd "$module_dir"; then
        GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=gcc \
        go build -buildmode=c-shared -ldflags="-s -w" -o "$output_dir" .
        
        if [ $? -ne 0 ]; then
            echo "[-] Build failed for module $module_name"
            cd - > /dev/null
            return 1
        fi

        cd - > /dev/null
        echo "[+] Done, saved in $output_dir"
    else
        echo "[-] Failed to cd into $module_dir"
        return 1
    fi
}


compile_windows_module() {
    module_name=$1
    module_dir="$modules_code_dir/$module_name"
    output_dir="$modules_compiled_dir/windows/$module_name.dll"
    echo "[+] Compiling Linux Module $module_name"

    if cd "$module_dir"; then
        GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
        go build -buildmode=c-shared -ldflags="-s -w" -o "$output_dir" .
        
        if [ $? -ne 0 ]; then
            echo "[-] Build failed for module $module_name"
            cd - > /dev/null
            return 1
        fi

        cd - > /dev/null
        echo "[+] Done, saved in $output_dir"
    else
        echo "[-] Failed to cd into $module_dir"
        return 1
    fi
}


compile_all_linux_module() {
    echo
    echo "###########################"
    echo "Compiling All Linux Modules"
    echo "###########################"
    echo

    for module in "${linuxmodules[@]}"; do
        compile_linux_module $module
    done

    echo
    echo "##########################"
    echo "Done compiling all modules"
    echo "##########################"
    echo
}


compile_all_windows_module() {
    echo
    echo "###########################"
    echo "Compiling All Windows Modules"
    echo "###########################"
    echo

    for module in "${windowsmodules[@]}"; do
        compile_windows_module $module
    done

    echo
    echo "##########################"
    echo "Done compiling all modules"
    echo "##########################"
    echo
}

print_help() {
    echo
    echo "Usage: $0 <command> [options]"
    echo
    echo "Commands:"
    echo "  single <linux|windows> <module>   Compile a single module"
    echo "  server-client                     Compile Orsted server and client"
    echo "  all                               Compile everything except protobuf"
    echo "  protobuf                          Compile protobuf files"
    echo "  help                              Show this help message"
    echo
    echo "Examples:"
    echo "  $0 single linux cat"
    echo "  $0 single windows psexec"
    echo "  $0 server-client"
    echo "  $0 all"
    echo "  $0 protobuf"
    echo
}

main() {
    case "$1" in
        single)
            type=$2
            module=$3
            if [[ -z "$type" || -z "$module" ]]; then
                echo "[-] Missing arguments"
                print_help
                return 1
            fi
            if [[ "$type" == "linux" ]]; then
                compile_linux_module "$module"
            elif [[ "$type" == "windows" ]]; then
                compile_windows_module "$module"
            else
                echo "[-] Unknown type: $type (must be linux or windows)"
                return 1
            fi
            ;;
        server-client)
            compile_client_server
            ;;
        all)
            compile_client_server
            compile_all_linux_module
            compile_all_windows_module
            ;;
        protobuf)
            compile_protobuf
            ;;
        help|"")
            print_help
            ;;
        *)
            echo "[-] Unknown command: $1"
            print_help
            return 1
            ;;
    esac
}

# Call main with all script arguments
main "$@"

