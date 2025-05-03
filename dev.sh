#!/bin/bash
# DocChain Development Script

# Set default values
PORT=${PORT:-8080}
DIFFICULTY=${DIFFICULTY:-4}
LOG_LEVEL=${LOG_LEVEL:-debug}
MODE=${MODE:-server}

# Print banner
echo "====================================="
echo "DocChain Development Script"
echo "====================================="
echo

# Function to show usage
show_usage() {
  echo "Usage: ./dev.sh [options]"
  echo
  echo "Options:"
  echo "  -m, --mode <mode>       Set mode (server, cli, build) [default: server]"
  echo "  -p, --port <port>       Set server port [default: 8080]"
  echo "  -d, --difficulty <num>  Set blockchain difficulty [default: 4]"
  echo "  -l, --log-level <level> Set log level (debug, info, warn, error) [default: debug]"
  echo "  -c, --clean             Clean build artifacts before running"
  echo "  -h, --help              Show this help message"
  echo
  echo "Examples:"
  echo "  ./dev.sh                          # Run server with default settings"
  echo "  ./dev.sh -m cli wallet create     # Run CLI command"
  echo "  ./dev.sh -p 9090 -d 5             # Run server on port 9090 with difficulty 5"
  echo "  ./dev.sh -m build -c              # Clean and build all binaries"
  echo
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -m|--mode)
      MODE="$2"
      shift 2
      ;;
    -p|--port)
      PORT="$2"
      shift 2
      ;;
    -d|--difficulty)
      DIFFICULTY="$2"
      shift 2
      ;;
    -l|--log-level)
      LOG_LEVEL="$2"
      shift 2
      ;;
    -c|--clean)
      CLEAN=true
      shift
      ;;
    -h|--help)
      show_usage
      exit 0
      ;;
    *)
      CLI_ARGS="$CLI_ARGS $1"
      shift
      ;;
  esac
done

# Clean if requested
if [ "$CLEAN" = true ]; then
  echo "Cleaning build artifacts..."
  make clean
fi

# Execute based on mode
case $MODE in
  server)
    echo "Starting server on port $PORT with difficulty $DIFFICULTY..."
    echo "Log level: $LOG_LEVEL"
    echo
    export PORT=$PORT
    export DIFFICULTY=$DIFFICULTY
    export LOG_LEVEL=$LOG_LEVEL
    make run
    ;;
  cli)
    echo "Running CLI command: $CLI_ARGS"
    echo
    make cli ARGS="$CLI_ARGS"
    ;;
  build)
    echo "Building all binaries..."
    echo
    make build
    ;;
  *)
    echo "Error: Unknown mode '$MODE'"
    show_usage
    exit 1
    ;;
esac
