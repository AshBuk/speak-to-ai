## 🖥️ Desktop Environment Support

> This document requires validation and testing across different environments. I have Fedora 42 (GNOME/Wayland).

> Open GitHub issues and community contributions can help make this project an excellent `speak to text` solution for the Linux community. 

### **For system tray on GNOME - to have full-featured UX with menu**:
```bash
# Ubuntu/Debian:
sudo apt install gnome-shell-extension-appindicator
# Fedora:
sudo dnf install gnome-shell-extension-appindicator
# Arch Linux:
sudo pacman -S gnome-shell-extension-appindicator
```
*KDE and other DEs have built-in system tray support, no need for appindicator*

### **Text output status (outputter, for automatic text insertion into active window)**

***Current Implementation: Smart Auto-Selection***
| Desktop Environment | Primary Tool | Fallback | Status |
|---------------------|--------------|----------|--------|
| **🟢 GNOME+Wayland** | ydotool | clipboard | ⚠️ Requires setup |
| **🟢 KDE+Wayland** | wtype → ydotool | clipboard | 🧪 Needs testing |
| **🟢 Sway/Other Wayland** | wtype → ydotool | clipboard | 🧪 Needs testing |
| **🟢 X11 (all DEs)** | xdotool | clipboard | ✅ Works out-of-box |

 *GNOME/Wayland requires ydotool setup. Other Wayland compositors may work with wtype without any setup - community testing needed*
 *RemoteDesktop Portal for GNOME/Wayland - Upcoming Feature!*

**Direct typing on Wayland - ydotool setup (recommended user-unit)**

> 1) Install ydotool:
```bash
sudo dnf install ydotool   # Fedora
sudo apt install ydotool   # Ubuntu/Debian
```
> 2) Allow access to /dev/uinput for non-root:
```bash
echo 'KERNEL=="uinput", GROUP="input", MODE="0660"' | sudo tee /etc/udev/rules.d/99-uinput.rules
sudo udevadm control --reload && sudo udevadm trigger
sudo usermod -a -G input $USER
# Re-login required for group change
```
> 3) Run ydotool as user-unit service (no root):
```bash
mkdir -p ~/.config/systemd/user
tee ~/.config/systemd/user/ydotool.service >/dev/null <<'EOF'
[Unit]
Description=ydotool user daemon

[Service]
ExecStart=/usr/bin/ydotoold --socket-perm=0660
Restart=always

[Install]
WantedBy=default.target
EOF
```
> 4) Restart and run the service
```bash
systemctl --user daemon-reload
systemctl --user enable --now ydotool
```
*This setup uses user service: safer and no root privileges needed*
*X11 works out-of-the-box without additional setups*

**Clipboard fallback**
- Works on **all** desktop environments  
- Requires manual `Ctrl+V` after speech recognition
- No additional setup needed

## ⌨️ **Hotkey Support Status (for hotkey registration and binding)**

### **AppImage Package**
- **All DEs:** **evdev first (requires input group)** → D-Bus GlobalShortcuts portal fallback
- **Optimization:** AppImage prioritizes evdev due to potential D-Bus portal limitations in sandboxed environment
- **Fallback:** If evdev unavailable, attempts D-Bus GlobalShortcuts portal
- **Setup input group:** 
```bash
sudo usermod -a -G input $USER
# Log out and log back in for changes to take effect
```

<!-- ### **Flatpak Package - upcoming feature!**
- **All DEs:** **D-Bus GlobalShortcuts portal only** (evdev blocked by sandbox security)
- **GNOME/KDE:** Works out-of-box via GlobalShortcuts portal
- **Other DEs:** Limited functionality if GlobalShortcuts portal unavailable
- **Experience:** Best on modern DEs with portal support -->

### **Alternative for Tiling WMs**
- **System hotkey tools:** sxhkd, xbindkeys, etc. + webhook integration

*Last updated: 2025-09-22*  
*Tested on: Fedora 42, Ubuntu 24.04*