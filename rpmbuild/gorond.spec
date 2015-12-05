%define _localbindir /usr/local/bin
%define _binaries_in_noarch_packages_terminate_build 0

Name:	  gorond
Version:	1.0.1
Release:	1
Summary:	custom cron powered by go.

Group:		uwork
License:	MIT
URL:		https://github.com/uwork/gorond
Source0:	%{name}.initd
Source1:  %{name}.sysconfig
Source2:  %{name}.conf
Packager: uwork

BuildArch:  noarch

%description
custom cron powered by go.

%prep


%build


%install
rm -rf %{buildroot}
install -d -m 755 %{buildroot}/%{_localbindir}
install    -m 655 %{_builddir}/%{name}   %{buildroot}/%{_localbindir}

install -d -m 755 %{buildroot}/%{_initrddir}
install    -m 755 %{_sourcedir}/%{name}.initd   %{buildroot}/%{_initrddir}/%{name}

install -d -m 755 %{buildroot}/%{_sysconfdir}/sysconfig/
install    -m 644 %{_sourcedir}/%{name}.sysconfig   %{buildroot}/%{_sysconfdir}/sysconfig/%{name}

install -d -m 755 %{buildroot}/%{_sysconfdir}/goron.d/
install    -m 644 %{_sourcedir}/%{name}.conf   %{buildroot}/%{_sysconfdir}/goron.conf

install -d -m 755 %{buildroot}/%{_logdir}/%{name}

%clean
rm -rf %{buildroot}%{_bindir}/%{name}

%pre

%post
chkconfig --add %{name}

%preun
if [ $1 -eq 0 ]; then
  service %{name} stop > /dev/null 2>&1
  chkconfig --del %{name}
fi

%files
%{_initrddir}/%{name}
%{_localbindir}/%{name}
%config(noreplace) %{_sysconfdir}/sysconfig/%{name}
%config(noreplace) %{_sysconfdir}/goron.conf
%dir %{_logdir}/%{name}
%dir %{_sysconfdir}/goron.d


%changelog
* Mon Nov 30 2015 uwork <github.com/uwork>
- first build.

