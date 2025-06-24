#
# spec file for package libdbusmenu
#
# Copyright (c) 2020 SUSE LLC
#
# All modifications and additions to the file contributed by third parties
# remain the property of their copyright owners, unless otherwise agreed
# upon. The license for this file, and modifications and additions to the
# file, is the same license as for the pristine package itself (unless the
# license for the pristine package is not an Open Source License, in which
# case the license is the MIT License). An "Open Source License" is a
# license that conforms to the Open Source Definition (Version 1.9)
# published by the Open Source Initiative.

# Please submit bugfixes or comments via https://bugs.opensuse.org/
#


%global flavor gtk2%{nil}
%global sname libappindicator
%define sname libdbusmenu
%define soname_glib 4
%define soname_gtk2 4
%define soname_gtk3 4
%define soname_jsonloader 4
%if "%{flavor}" == "gtk2"
%global gtkver 2
%global soname_gtk %{soname_gtk2}
%global libname_gtk  libdbusmenu-gtk%{soname_gtk}
%global psuffix      -gtk%{gtkver}
# dumper is GTK2 only
%ifarch %arm
# Valgrind is broken on armv6/7 atm
%bcond_with testtools
%else
%bcond_without testtools
%endif
%global package_glib 1
# Docs are the same for GTK2/3, dito for glib
%bcond_without docs
%endif
%if "%{flavor}" == "gtk3"
%global gtksuffix 3
%global gtkver 3
%global soname_gtk %{soname_gtk3}
%global libname_gtk  libdbusmenu-gtk3-%{soname_gtk}
%global psuffix      -gtk%{gtkver}
%bcond_with    testtools
%bcond_with    docs
%endif
%global libname_glib libdbusmenu-glib%{soname_glib}
Name:           libdbusmenu%{?psuffix}
Version:        16.04.0
Release:        150200.3.2.1
Summary:        Small library that passes a menu structure across DBus
License:        GPL-3.0-only AND (LGPL-2.1-only OR LGPL-3.0-only)
Group:          System/Libraries
URL:            https://launchpad.net/dbusmenu
Source:         https://launchpad.net/libdbusmenu/16.04/%{version}/+download/%{sname}-%{version}.tar.gz
# PATCH-FIX-OPENSUSE
Patch0:         0001-Fix-build-with-gtk-doc-1.32-due-to-non-existing-tree.patch
# PATCH-FIX-UPSTREAM
Patch1:         0002-genericmenuitem-Make-accelerator-text-appear-again.patch
# PATCH-FIX-OPENSUSE
Patch2:         0003-Fix-HAVE_VALGRIND-AM_CONDITIONAL.patch
BuildRequires:  autoconf
BuildRequires:  automake
BuildRequires:  intltool
BuildRequires:  libtool
BuildRequires:  pkgconfig
BuildRequires:  vala
BuildRequires:  pkgconfig(atk)
BuildRequires:  pkgconfig(dbus-glib-1)
BuildRequires:  pkgconfig(gdk-pixbuf-2.0)
BuildRequires:  pkgconfig(gobject-introspection-1.0)
%if "%{flavor}" == ""
ExclusiveArch:  do-not-build
%endif
%if %{with docs}
BuildRequires:  gtk-doc
BuildRequires:  pkgconfig(gnome-doc-utils)
%endif
%if "%flavor" == "gtk2"
BuildRequires:  pkgconfig(gtk+-2.0)
%else
BuildRequires:  pkgconfig(gtk+-3.0)
%endif
%if %{with testtools}
BuildRequires:  pkgconfig(json-glib-1.0)
BuildRequires:  pkgconfig(valgrind)
BuildRequires:  pkgconfig(x11)
%endif

%description
A small little library that was created by pulling out some common
code out of mate-indicator-applet. It passes a menu structure
across D-Bus so that a program can create a menu simply without
worrying about how it is displayed on the other side of the bus.

%package -n libdbusmenu-tools
Summary:        Development tools for the dbusmenu libraries
Group:          Development/Tools/Other
Requires:       %{libname_glib} = %{version}

%description -n libdbusmenu-tools
This packages contains the development tools for the dbusmenu libraries.

%package -n %{libname_glib}
Summary:        Small library that passes a menu structure across D-Bus
Group:          System/Libraries

%description -n %{libname_glib}
This package contains the shared library for the dbusmenu-glib.

%package -n typelib-1_0-Dbusmenu-0_4
Summary:        Introspection bindings for %{libname_glib}
Group:          System/Libraries

%description -n typelib-1_0-Dbusmenu-0_4
This package contains the GObject Introspection bindings for the dbusmenu
library.

%package -n libdbusmenu-glib-devel
Summary:        Development files for libdbusmenu-glib
Group:          Development/Libraries/C and C++
Requires:       %{libname_glib} = %{version}
Requires:       typelib-1_0-Dbusmenu-0_4 = %{version}
Requires:       pkgconfig(dbus-glib-1)

%description -n libdbusmenu-glib-devel
This package contains the development files for the dbusmenu-glib library.

%package -n libdbusmenu-glib-doc
Summary:        Documentation for libdbusmenu-glib%{soname_glib}
Group:          Documentation/HTML
BuildArch:      noarch

%description -n libdbusmenu-glib-doc
This package includes the documentation for the dbusmenu-glib library.

%package -n %{libname_gtk}
Summary:        GTK+ %{gtkver} version of libdbusmenu
Group:          System/Libraries
%if "%{flavor}" == "gtk2"
Requires:       gtk2
%endif

%description -n %{libname_gtk}
This package contains GTK %{gtkver} dbusmenu shared library.

%package -n typelib-1_0-DbusmenuGtk%{?gtksuffix}-0_4
Summary:        Introspection bindings for %{libname_gtk}
Group:          System/Libraries

%description -n typelib-1_0-DbusmenuGtk%{?gtksuffix}-0_4
This package contains the GObject Introspection bindings for the GTK+ %{gtkver} version
of the dbusmenu-gtk library.

%package devel
Summary:        Development files for %{libname_gtk}
Group:          Development/Libraries/C and C++
Requires:       %{libname_gtk} = %{version}
Requires:       typelib-1_0-DbusmenuGtk%{?gtksuffix}-0_4
Requires:       pkgconfig(dbus-glib-1)
Requires:       pkgconfig(dbusmenu-glib-0.4) = %{version}

%description devel
This package contains the development files for the dbusmenu-gtk%{gtkver} library.

%package doc
Summary:        Documentation for libdbusmenu - GTK 2 and GTK 3
Group:          Documentation/HTML
BuildArch:      noarch

%description doc
This package contains the documentation for the dbusmenu-gtk2 and dbusmenu-gtk3
libraries.

%package -n libdbusmenu-jsonloader%{soname_jsonloader}
Summary:        Small library that passes a menu structure across DBus -- Test library
Group:          System/Libraries

%description -n libdbusmenu-jsonloader%{soname_jsonloader}
This package contains the shared libraries for dbusmenu-jsonloader, a library
meant for test suites.

%package -n libdbusmenu-jsonloader-devel
Summary:        Development files for libdbusmenu-jsonloader%{soname_jsonloader}
Group:          Development/Libraries/C and C++
Requires:       libdbusmenu-jsonloader%{soname_jsonloader} = %{version}
Requires:       pkgconfig(dbus-glib-1)
Requires:       pkgconfig(dbusmenu-glib-0.4) = %{version}
Requires:       pkgconfig(json-glib-1.0)

%description -n libdbusmenu-jsonloader-devel
This package contains the development files for the dbusmenu-jsonloader library.

%prep
%setup -q -n %{sname}-%{version}
%patch0 -p1
%patch1 -p1
%patch2 -p1

%build
export CFLAGS="%{optflags} -Wno-error"
autoreconf -vfi

%configure \
        --disable-static       \
%if 0%{without testtools}
        --disable-dumper       \
        --disable-tests        \
%endif
        --enable-introspection \
        --with-gtk=%{gtkver}

make %{?_smp_mflags}

%install
%make_install

find %{buildroot} -type f -name "*.la" -delete -print

%if %{with testtools}
# Put documentation in correct directory.
mkdir -p %{buildroot}%{_docdir}/%{sname}-tools/
mv -f %{buildroot}%{_datadir}/doc/%{sname}/README.dbusmenu-bench \
  %{buildroot}%{_docdir}/%{sname}-tools/

%else
# Cleanup unwanted files
rm -Rf %{buildroot}%{_datadir}/doc/%{sname}/README.dbusmenu-bench \
rm -Rf %{buildroot}%{_datadir}/%{sname}
rm -Rf %{buildroot}%{_libexecdir}/dbusmenu-{bench,dumper,testapp}

%endif

# Remove glib version (only package once)
%if 0%{?package_glib}
# Put examples in correct documentation directory.
%if %{with testtools}
mkdir -p %{buildroot}%{_docdir}/%{sname}-glib-devel/examples/
mv %{buildroot}%{_datadir}/doc/%{sname}/examples/glib-server-nomenu.c \
  %{buildroot}%{_docdir}/%{sname}-glib-devel/examples/
%endif

%else
rm -Rf %{buildroot}%{_includedir}/libdbusmenu-glib-0.4/
rm -Rf %{buildroot}%{_libdir}/libdbusmenu-glib.so*
rm -Rf %{buildroot}%{_libdir}/pkgconfig/dbusmenu-glib-0.4.pc
rm -Rf %{buildroot}%{_libdir}/girepository-1.0/Dbusmenu-0.4.typelib
rm -Rf %{buildroot}%{_datadir}/gir-1.0/Dbusmenu-0.4.gir
rm -Rf %{buildroot}%{_datadir}/vala/vapi/Dbusmenu-0.4.vapi
%endif

%if %{without docs}
# (Bundled) docs are installed even with --disable-gtk-doc
rm -Rf %{buildroot}%{_datadir}/gtk-doc
%endif

%post -n %{libname_glib} -p /sbin/ldconfig
%postun -n %{libname_glib} -p /sbin/ldconfig
%post -n %{libname_gtk} -p /sbin/ldconfig
%postun -n %{libname_gtk} -p /sbin/ldconfig
%post -n libdbusmenu-jsonloader%{soname_jsonloader} -p /sbin/ldconfig
%postun -n libdbusmenu-jsonloader%{soname_jsonloader} -p /sbin/ldconfig

%if %{with testtools}
%files -n libdbusmenu-tools
%license COPYING*
%doc NEWS
%{_libexecdir}/dbusmenu-bench
%{_libexecdir}/dbusmenu-dumper
%{_libexecdir}/dbusmenu-testapp
%dir %{_datadir}/%{sname}/
%dir %{_datadir}/%{sname}/json/
%{_datadir}/%{sname}/json/test-gtk-label.json
%doc %dir %{_docdir}/%{sname}-tools/
%doc %{_docdir}/%{sname}-tools/README.dbusmenu-bench
%endif

%if 0%{?package_glib}
%files -n %{libname_glib}
%license COPYING*
%doc NEWS
%{_libdir}/libdbusmenu-glib.so.%{soname_glib}*

%files -n typelib-1_0-Dbusmenu-0_4
%license COPYING*
%doc NEWS
%{_libdir}/girepository-1.0/Dbusmenu-0.4.typelib

%files -n libdbusmenu-glib-devel
%license COPYING*
%doc NEWS
%dir %{_includedir}/libdbusmenu-glib-0.4/
%dir %{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/client.h
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/dbusmenu-glib.h
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/enum-types.h
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/menuitem-proxy.h
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/menuitem.h
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/server.h
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-glib/types.h
%{_libdir}/pkgconfig/dbusmenu-glib-0.4.pc
%{_libdir}/libdbusmenu-glib.so
%{_datadir}/gir-1.0/Dbusmenu-0.4.gir
%dir %{_datadir}/vala/vapi/
%{_datadir}/vala/vapi/Dbusmenu-0.4.vapi
%if %{with testtools}
%doc %dir %{_docdir}/%{sname}-glib-devel/examples/
%doc %{_docdir}/%{sname}-glib-devel/examples/glib-server-nomenu.c
%endif

%files -n libdbusmenu-glib-doc
%license COPYING*
%doc NEWS
%doc %{_datadir}/gtk-doc/html/libdbusmenu-glib/
%endif

%files -n %{libname_gtk}
%license COPYING*
%doc NEWS
%{_libdir}/libdbusmenu-gtk*.so.%{soname_gtk}*

%files -n typelib-1_0-DbusmenuGtk%{?gtksuffix}-0_4
%license COPYING*
%doc NEWS
%{_libdir}/girepository-1.0/DbusmenuGtk*-0.4.typelib

%files devel
%license COPYING*
%doc NEWS
%dir %{_includedir}/libdbusmenu-gtk*-0.4/
%dir %{_includedir}/libdbusmenu-gtk*-0.4/libdbusmenu-gtk/
%{_includedir}/libdbusmenu-gtk*-0.4/libdbusmenu-gtk/client.h
%{_includedir}/libdbusmenu-gtk*-0.4/libdbusmenu-gtk/dbusmenu-gtk.h
%{_includedir}/libdbusmenu-gtk*-0.4/libdbusmenu-gtk/menu.h
%{_includedir}/libdbusmenu-gtk*-0.4/libdbusmenu-gtk/menuitem.h
%{_includedir}/libdbusmenu-gtk*-0.4/libdbusmenu-gtk/parser.h
%{_libdir}/pkgconfig/dbusmenu-gtk*-0.4.pc
%{_libdir}/libdbusmenu-gtk*.so
%{_datadir}/gir-1.0/DbusmenuGtk*-0.4.gir
%dir %{_datadir}/vala/vapi/
%{_datadir}/vala/vapi/DbusmenuGtk*-0.4.vapi

%if %{with docs}
%files doc
%license COPYING*
%doc NEWS
%doc %{_datadir}/gtk-doc/html/libdbusmenu-gtk/
%endif

%if %{with testtools}
%files -n libdbusmenu-jsonloader%{soname_jsonloader}
%license COPYING*
%doc NEWS
%{_libdir}/libdbusmenu-jsonloader.so.%{soname_jsonloader}*

%files -n libdbusmenu-jsonloader-devel
%license COPYING*
%doc NEWS
%dir %{_includedir}/libdbusmenu-glib-0.4/
%dir %{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-jsonloader/
%{_includedir}/libdbusmenu-glib-0.4/libdbusmenu-jsonloader/json-loader.h
%{_libdir}/pkgconfig/dbusmenu-jsonloader-0.4.pc
%{_libdir}/libdbusmenu-jsonloader.so
%endif

%changelog
* Mon Mar 23 2020 dimstar@opensuse.org
- Require the tyeplib packages from the -devel packages: typelibs
  are shared libraries and consumers of the -devel package have a
  right to assume the libraries are present.
* Wed Dec 11 2019 guillaume.gardet@opensuse.org
- Disable testtools on %%arm since Valgrind is broken on armv6/7 atm
* Tue Nov 19 2019 stefan.bruens@rwth-aachen.de
- Work around OBS idiosyncrasies regarding packages name.
* Mon Nov 18 2019 stefan.bruens@rwth-aachen.de
- Drop dependency on deprecated gnome-common, just run autoreconf
- Do not include unused tree_index.sgml, fix build with gtk-doc >= 1.32,
  see https://gitlab.gnome.org/GNOME/gtk-doc/issues/103
  * add 0001-Fix-build-with-gtk-doc-1.32-due-to-non-existing-tree.patch
- Fix missing accelerators, add
  0002-genericmenuitem-Make-accelerator-text-appear-again.patch
- Split Gtk2 and Gtk3 build - glib, tools and doc subpackage are created
  from the Gtk2 flavor.
  * Fix building with disabled tests, add 0003-Fix-HAVE_VALGRIND-AM_CONDITIONAL.patch
* Wed Oct 16 2019 dimstar@opensuse.org
- Inject -Wno-error into CFLAGS. It's kinda ridiculous for code
  that is not maintained upstream to pass -Werror by default and
  then not catching up. So for now we accept warnings.
* Tue Aug 13 2019 bjorn.lie@gmail.com
- Drop superfluous hard pkgconfig(gtk+-2.0) Requires from
  libdebusmenu-glib-devel sub-package.
* Wed Feb 28 2018 dimstar@opensuse.org
- Modernize spec-file by calling spec-cleaner
* Thu Sep  8 2016 dimstar@opensuse.org
- Use proper libdbusmenu 16.04.0 tarball directly from launchpad:
  + The old tarball's configure.ac happened to still use an old
    version, causing the .pc file to advertise insufficient
    capabilities.
* Fri Mar  4 2016 sor.alexei@meowr.ru
- Update to 12.10.3+bzr20160223 (changes since 12.10.3+bzr20150410):
  * Disable test-json-instruction, hangs on builds (lp#1429291).
  * gtk: Look for GtkImages on regular GtkMenuItems too (lp#1549021).
* Tue May  5 2015 sor.alexei@meowr.ru
- Update to 12.10.3+bzr20150410 (changes since 12.10.3+bzr20140610):
  * Use the configure-generated libtool script instead of
    /usr/bin/libtool, which might not match what we have.
  * Use gi's typelibdir pkgconfig variable and install into this
    directory, now that gi supports multiarch.
  * Parser: don't override the label for stock items if a custom
    one is provided.
- Minor spec cleanup.
* Sun Oct 26 2014 p.drouand@gmail.com
- Update to version 12.10.3+14.10.20140610
  + No changelog available
* Mon Mar 10 2014 cfarrell@suse.com
- license update: GPL-3.0 and (LGPL-2.1 or LGPL-3.0)
  Interaction is not aggregation (^and^) but rather option (^or^). Also,
  the GPL-3.0 components are in the separate tools/ subdirectory
* Fri Mar  7 2014 hrvoje.senjan@gmail.com
- Init libdbusmenu package
