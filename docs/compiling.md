## Linux

**Dependencies**
```bash
# Debian/Ubuntu
sudo apt install opencl-headers ocl-icd-opencl-dev build-essential

# Arch
sudo pacman -S opencl-headers ocl-icd gcc
```

**Enable CGO and build**
```bash
go env -w CGO_ENABLED=1
go build
```

**GPU runtime drivers**

For NVIDIA:
```bash
# Debian/Ubuntu
sudo apt install nvidia-opencl-icd

# Arch
sudo pacman -S opencl-nvidia
```

For AMD (RDNA / GCN):
```bash
# Debian/Ubuntu — ROCm OpenCL
sudo apt install rocm-opencl-runtime

# Arch
sudo pacman -S opencl-amd   # or rocm-opencl-runtime from AUR
```

> [!NOTE]
> AMD GPU support requires the ROCm or AMDGPU-PRO OpenCL runtime. The open-source
> Mesa `rusticl` driver may also work but has not been tested.

---

## Windows

You need the [Zig](https://ziglang.org/download/) compiler (any recent release) as the C cross-compiler.

```powershell
$env:CGO_ENABLED="1"
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CC="zig cc -target x86_64-windows"
$env:CXX="zig c++ -target x86_64-windows"
go build
```

For GPU support on Windows install the vendor OpenCL runtime:
- **NVIDIA**: install the standard Game Ready / Studio driver (includes OpenCL)
- **AMD**: install AMD Software: Adrenalin Edition (includes OpenCL)

---

## Cross-compile Linux → Windows

```bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC="zig cc -target x86_64-windows" \
  CXX="zig c++ -target x86_64-windows" \
  go build
```
