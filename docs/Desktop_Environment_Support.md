## 🖥️ Desktop Environment Support

> Tested on major Linux distributions (Fedora, Ubuntu, Arch Linux) and tiling window managers (Hyprland, Sway)

> If you encounter issues with your desktop environment, feel free to [open an issue](https://github.com/asherbuk/speak-to-ai/issues). 

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

**Current Implementation: Smart Auto-Selection**

| Desktop Environment | Primary Tool | Fallback | Status |
|---------------------|--------------|----------|--------|
| **🟢 GNOME+Wayland** | ydotool | clipboard | ⚠️ Requires setup |
| **🟢 KDE+Wayland** | wtype → ydotool | clipboard | ✅ Auto-detected |
| **🟢 Sway/Other Wayland** | wtype → ydotool | clipboard | ✅ Auto-detected |
| **🟢 X11 (all DEs)** | xdotool | clipboard | ✅ Works out-of-box |

 *GNOME/Wayland requires ydotool setup. Other Wayland compositors (KDE, Sway, etc.) may work with wtype without any setup*
 *RemoteDesktop Portal for GNOME/Wayland - Upcoming Feature!*

## Direct typing on Wayland - Tool options

The application automatically selects the best available typing tool:
- **wtype**: Works without setup on non-GNOME Wayland compositors (KDE, Sway, etc.). Automatically selected if available.
- **ydotool**: Required for GNOME/Wayland, also works as fallback on other Wayland compositors. Requires setup (see below).

### ydotool setup (recommended user-unit)

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
*For non-GNOME Wayland compositors, wtype may work without any setup - the app will automatically try it first*

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

### **Alternative for Tiling WMs**
- **CLI commands:** Direct hotkey binding via WM config (i3, sway, bspwm, etc.)
- **Example (i3/sway):**
  ```
  bindsym $mod+r exec speak-to-ai start
  bindsym $mod+Shift+r exec speak-to-ai stop
  ```
- See [CLI Usage Guide](CLI_USAGE.md) for command reference

*Last updated: 2025-12-21*  
*Tested on: Fedora 42, Ubuntu 24.04*