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

 *wtype doesn't work on GNOME/Wayland - compositor limitation, so use clipboard (ctrl+v) or setup ydotool*
 *RemoteDesktop Portal for GNOME/Wayland - Upcoming Feature!*

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

### **Native System Install**
- **GNOME/KDE (Wayland/X11):** D-Bus GlobalShortcuts portal → evdev fallback (requires `input` group)
- **Other DEs (XFCE/MATE/LXQt):** D-Bus GlobalShortcuts portal → evdev fallback (requires `input` group)
- **Tiling WMs (i3/sway/dwm):** evdev only (requires `input` group)
- **Experience:** Modern DEs with portal support work without setup, others require `input` group

### **AppImage Package**
- **All DEs:** **evdev first** → D-Bus GlobalShortcuts portal fallback
- **Optimization:** AppImage prioritizes evdev due to potential D-Bus portal limitations in sandboxed environment
- **Setup Required:** `sudo usermod -a -G input $USER` + logout/login for reliable operation
- **Fallback:** If evdev unavailable, attempts D-Bus GlobalShortcuts portal

### **Flatpak Package**
- **All DEs:** **D-Bus GlobalShortcuts portal only** (evdev blocked by sandbox security)
- **GNOME/KDE:** Works out-of-box via GlobalShortcuts portal
- **Other DEs:** Limited functionality if GlobalShortcuts portal unavailable
- **Experience:** Best on modern DEs with portal support

### **Alternative for Tiling WMs**
- **System hotkey tools:** sxhkd, xbindkeys, etc. + webhook integration

*Last updated: 2025-09-16*  
*Tested on: Fedora 42, Ubuntu 24.04*