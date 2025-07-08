# Desktop Surveillance Camera

A Windows desktop monitoring program written in Go that allows remote screen monitoring through a web interface.

## Features

- **Windows Screenshot**: Efficient screen capture using Win32 API
- **Web Interface**: HTML interface accessible from mobile browsers
- **Dual Operation Modes**:
  - On-demand mode: Captures screenshots only when accessed, saving resources
  - Real-time mode: Automatically captures screenshots at regular intervals for live updates
- **JavaScript-free Compatible**: Supports environments with JavaScript disabled
- **LAN Deployment**: Perfect for home/office network use

## Quick Start

### 1. Build the Program

**Using Makefile (Recommended)**:

```bash
# Build Windows version
make windows

# Build 32-bit Windows version  
make windows-386

# View all available commands
make help
```

**Manual Build**:

```bash
go build -o surveillance-camera.exe .
```

### 2. Run the Program

```bash
# Run with default configuration
surveillance-camera.exe

# Test screenshot functionality
surveillance-camera.exe -test

# Use custom configuration
surveillance-camera.exe -config my-config.json

# Show help
surveillance-camera.exe -help
```

### 3. Access the Interface

After starting, open in browser: `http://your-computer-ip:9981`

## Configuration

The program automatically creates a default configuration file `config.json`:

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 9981
  },
  "capture": {
    "mode": "ondemand",
    "interval": "5s"
  }
}
```

### Configuration Options

- `server.host`: Server listening address (0.0.0.0 means all network interfaces)
- `server.port`: Server port
- `capture.mode`: Screenshot mode
  - `ondemand`: On-demand mode, captures only when accessed
  - `realtime`: Real-time mode, captures automatically at intervals
- `capture.interval`: Screenshot interval for real-time mode (e.g., "5s", "10s", "1m")

## API Endpoints

- `GET /`: Main page (HTML interface)
- `GET /last`: Get latest screenshot (PNG format)

## Use Cases

1. **Remote Monitoring**: View home computer screen from mobile
2. **Office Supervision**: Monitor work computer usage
3. **System Status**: Remote desktop status checking for servers
4. **Presentation Aid**: Real-time screen sharing

## System Requirements

- Windows 7/8/10/11
- Go 1.18+ (only required for compilation)
- CGO-enabled compilation environment

## Security Recommendations

- Use only in trusted networks
- Configure firewall to restrict access
- Consider using non-default ports
- Add basic authentication if needed (can be extended)

## Troubleshooting

1. **Screenshot Failure**: Ensure program runs with sufficient privileges
2. **Access Issues**: Check firewall settings and port conflicts
3. **Performance Issues**: Adjust screenshot interval in real-time mode
4. **Compilation Errors**: Ensure CGO environment is properly configured

## Build Instructions

This project uses CGO to call Windows APIs and requires a CGO environment:

**Building on Windows**:

- Install TDM-GCC or MinGW-w64
- Ensure gcc is in PATH
- Run `make windows`

**Cross-compiling from Linux to Windows**:

- Install mingw-w64 cross-compilation toolchain
- Ubuntu: `sudo apt-get install gcc-mingw-w64`
- Set environment variable: `export CC=x86_64-w64-mingw32-gcc`
- Run `make windows`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Support

If you encounter any issues or have questions, please:

1. Check the troubleshooting section above
2. Search existing issues on GitHub
3. Create a new issue with detailed information

## Acknowledgments

- Thanks to the Go community for excellent cross-platform support
- Windows API documentation and examples
- All contributors and users of this project