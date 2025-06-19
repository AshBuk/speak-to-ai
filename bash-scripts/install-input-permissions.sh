#!/bin/bash

# Script to set up input device permissions for speak-to-ai
# This allows the application to access keyboard events for hotkeys

echo "Setting up input device permissions for Speak-to-AI..."

# Create udev rule for input devices
UDEV_RULE="/etc/udev/rules.d/99-speak-to-ai-input.rules"

echo "Creating udev rule at: $UDEV_RULE"
sudo tee "$UDEV_RULE" > /dev/null << 'EOF'
# Allow users in input group to access input devices
# This is needed for speak-to-ai hotkey functionality
KERNEL=="event*", GROUP="input", MODE="0660"
SUBSYSTEM=="input", GROUP="input", MODE="0660"
EOF

# Add current user to input group
echo "Adding current user ($USER) to input group..."
sudo usermod -a -G input "$USER"

# Reload udev rules
echo "Reloading udev rules..."
sudo udevadm control --reload-rules
sudo udevadm trigger

echo "✅ Input permissions configured successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Log out and log back in (or reboot) for group changes to take effect"
echo "2. Run speak-to-ai - hotkeys should now work"
echo ""
echo "🔧 Alternative: run speak-to-ai with sudo for immediate testing:"
echo "   sudo ./speak-to-ai"
echo ""
echo "⚠️  Note: The udev rule allows all users in 'input' group to access input devices."
echo "   This is generally safe but removes some security isolation." 