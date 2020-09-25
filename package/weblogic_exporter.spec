# Avoid build ID error due to use of precompiled binaries
%define debug_package %{nil}

Name:           weblogic_exporter
Version:        1.0.0
Release:        1%{?dist}
Summary:        A Prometheus exporter for Oracle Weblogic Server.
License:        Apache 2.0
URL:            https://github.com/benridley/wls_go.git
Source0:        weblogic_exporter-%{version}.tar.gz
BuildArch:      x86_64
BuildRequires:  go
Requires(pre): shadow-utils

#%define _unpackaged_files_terminate_build 0

%description
A Prometheus exporter for Oracle Weblogic Server.

%prep
%setup -q %{SOURCE0} -n weblogic_exporter-%{version}

%build
go build -o bin/weblogic_exporter -mod vendor src/main.go

%install
mkdir -p %{buildroot}/opt/weblogic_exporter
mkdir -p %{buildroot}/usr/lib/systemd/system
install --mode=770 bin/weblogic_exporter %{buildroot}/opt/weblogic_exporter
install --mode=660 config.yaml %{buildroot}/opt/weblogic_exporter
install --mode=644 package/weblogic_exporter.service %{buildroot}/usr/lib/systemd/system/weblogic_exporter.service

%pre
mkdir -p /opt/weblogic_exporter
getent group weblogic_exporter >/dev/null || groupadd -r weblogic_exporter
getent passwd weblogic_exporter >/dev/null || useradd -g weblogic_exporter -d /opt/weblogic_exporter -s /sbin/nologin -c "Prometheus Weblogic Exporter System Account" weblogic_exporter
exit 0

%post

%files
/usr/lib/systemd/system/weblogic_exporter.service
%attr(-, weblogic_exporter, weblogic_exporter) /opt/weblogic_exporter
%config /opt/weblogic_exporter/config.yaml


%changelog
* Fri Jun 19 2020 Ben Ridley - %{version}-%{release}
- Add systemd unit file
* Thu Jun 18 2020 Ben Ridley - %{version}-%{release}
- Initial Creation
