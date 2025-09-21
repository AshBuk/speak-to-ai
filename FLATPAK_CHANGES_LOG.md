# Flatpak Manifest Changes Log

- 2025-09-21: Add global PKG_CONFIG_PATH in build-options to expose /app pkgconfig dirs during builds.
- 2025-09-21: Harden ayatana-ido pkg-config placement: ensure /app/lib/pkgconfig/ayatana-ido3-0.4.pc exists and create symlink libayatana-ido3-0.4.pc.
- 2025-09-21: ydotool: remove scdoc module; disable manpages via post-extract (comment add_subdirectory(manpage), replace scdoc with true); keep append-path minimal.
- 2025-09-21: ydotool: hard-disable manpage generation. Post-extract: comment out all add_subdirectory(manpage) patterns across all CMakeLists and stub manpage/CMakeLists.txt. Revert to scdoc stub via /usr/bin/true due to upstream lacking meson.build in tag 1.11.2.
 - 2025-09-21: ydotool: switch to simple build; run cmake+ninja manually; force replace any generated scdoc calls under build/ to /app/bin/scdoc (stub).
 - 2025-09-21: Dockerfile.flatpak: install newer flatpak/flatpak-builder from PPA flatpak/stable to ensure appstream-compose/appstreamcli compose available during export.
