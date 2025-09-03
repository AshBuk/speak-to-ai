## 🖥️ Desktop Environment Support

>take 1: This document requires validation and testing across different environments. I have Fedora 42 (GNOME/Wayland).
>take 2: Open GitHub issues and community contributions can help make this project an excellent solution for the Linux community. 

### 🟢 **Excellent Support**

#### **GNOME (Wayland/X11)**
- **Distributions:** Ubuntu 22.04+, Fedora 36+, openSUSE Tumbleweed
- **AppImage:** evdev (requires `sudo usermod -a -G input $USER && reboot`)
- **Flatpak:** D-Bus GlobalShortcuts portal (works out of the box)
- **Experience:** Hotkeys work immediately after setup

#### **KDE Plasma (Wayland/X11)**  
- **Distributions:** KDE Neon, Kubuntu, Manjaro KDE, openSUSE
- **AppImage:** evdev (requires `sudo usermod -a -G input $USER && reboot`)
- **Flatpak:** D-Bus GlobalShortcuts + KGlobalAccel (works out of the box)
- **Experience:** Same as GNOME - excellent with proper setup

### 🟡 **Good Support**

#### **XFCE**
- **Distributions:** Xubuntu, Xfce spin of Fedora/openSUSE
- **AppImage:** evdev (requires `sudo usermod -a -G input $USER && reboot`)
- **Flatpak:** Limited D-Bus support, may require manual setup
- **Experience:** Requires `input` group membership

#### **MATE**
- **Distributions:** Ubuntu MATE, Fedora MATE spin
- **AppImage:** evdev (requires `sudo usermod -a -G input $USER && reboot`)
- **Flatpak:** Limited portal support, evdev fallback available

#### **LXQt/LXDE**
- **Distributions:** Lubuntu, PCLinuxOS
- **AppImage:** evdev (requires `sudo usermod -a -G input $USER && reboot`)
- **Flatpak:** No portal support, evdev required
- **Note:** Lightweight DEs typically lack portal infrastructure

### 🔴 **Limited Support**

#### **Tiling Window Managers**
- **Examples:** i3, sway, dwm, awesome, bspwm
- **AppImage:** evdev only (requires `sudo usermod -a -G input $USER && reboot`)
- **Flatpak:** No portal support, evdev required with input devices permission
- **Requirement:** User MUST be in `input` group
- **Alternative:** Manual system hotkey setup (sxhkd, etc.)

---

*Last updated: 2025-09-03*  
*Tested on: Fedora 42, Ubuntu 24.04,*