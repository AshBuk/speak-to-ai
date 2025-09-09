## 🖥️ Desktop Environment Support

> This document requires validation and testing across different environments. I have Fedora 42 (GNOME/Wayland).

> Open GitHub issues and community contributions can help make this project an excellent `speak to text` solution for the Linux community. 

### **For system tray on GNOME - to have full-featured UX with menu**:
```bash
# Ubuntu/Debian
sudo apt install gnome-shell-extension-appindicator
# Fedora
sudo dnf install gnome-shell-extension-appindicator
# Arch Linux
sudo pacman -S gnome-shell-extension-appindicator
```
*KDE and other DEs have built-in system tray support, no need for appindicator*

### **Text output status (outputter, for automatic insertion of speech-to-text into active window)**

***Current Implementation: Smart Auto-Selection***
| Desktop Environment | Primary Tool | Fallback | Status |
|---------------------|--------------|----------|--------|
| **🟢 GNOME+Wayland** | ydotool* | clipboard | ⚠️ Requires setup |
| **🟢 KDE+Wayland** | wtype → ydotool → xdotool | clipboard | ✅ Works out-of-box |
| **🟢 Sway/i3** | wtype → ydotool → xdotool | clipboard | ✅ Works out-of-box |
| **🟢 X11 (all DEs)** | xdotool | clipboard | ✅ Works out-of-box |

*wtype doesn't work on GNOME/Wayland - compositor limitation, so use clipboard (ctrl+v) or setup ydotool
*RemoteDesktop Portal for GNOME/Wayland - Upcoming Feature!

**Outputter setup - ydotool (requires for GNOME!)**
```bash
# Fedora
sudo dnf install ydotool
# Debian/Ubuntu-based
sudo apt install ydotool
# Add to input group
sudo usermod -a -G input $USER            
# logout → login (or reboot), then:
sudo systemctl enable --now ydotoold      # Start daemon
```

**Clipboard fallback**
- Works on **all** desktop environments  
- Requires manual `Ctrl+V` after speech recognition
- No additional setup needed

## ⌨️ **Hotkey Support Status (for hotkey registration and binding)**

### **GNOME (Wayland/X11)**
- **AppImage:** D-Bus portal → evdev fallback (requires `input` group)
- **Flatpak:** D-Bus portal → evdev fallback (may require `input` group)
- **Experience:** Portal usually works, evdev fallback if portal unavailable

### **KDE Plasma (Wayland/X11)**  
- **AppImage:** D-Bus portal → evdev fallback (requires `input` group)
- **Flatpak:** D-Bus portal → evdev fallback (may require `input` group)
- **Experience:** Better portal support than GNOME, but fallbacks available

### **Other DEs (XFCE/MATE/LXQt)**
- **AppImage:** evdev only (requires `input` group)
- **Flatpak:** Limited portal support, evdev fallback available
- **Experience:** Requires `input` group membership for reliable hotkeys

### **Tiling WMs (i3/sway/dwm/bspwm)**
- **AppImage:** evdev only (requires `input` group)
- **Flatpak:** evdev only (requires `input` group)
- **Alternative:** System hotkey tools (sxhkd, etc.) + webhook integration

*Last updated: 2025-09-05*  
*Tested on: Fedora 42, Ubuntu 24.04*