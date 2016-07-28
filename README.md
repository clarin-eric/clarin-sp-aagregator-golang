# clarin-sp-aagregator-golang
This project is a cgi-scipt, implemented in golang, with the aim to collect and send shibbolleth attribute information to the CLARIN aagregator.

The implementation is a port of the php based [clarin-sp-aaggregator](https://github.com/ufal/clarin-sp-aaggregator).

# Configuration

## Shibboleth Service Provider
Follow the guidelines provided [here](https://github.com/ufal/clarin-sp-aaggregator#shibboleth2xml).

Use `sessionHook="/aa-statistics"` instead of `sessionHook="/php/aa-statistics.php"`.

## Apache web server

Put the golang executable somewhere in your document root. Assuming the document root is `/var/www/html` and we want to expose the hook under `/hook`, the following location shoulw be used:

```
/var/www/html/aa-statistics/clarin-sp-aagregator-golang.go
```

Enable mod_cgi and restart apache:

```
a2enmod cgi
apachectl restart
```

Configure the golang cgi executable on a location in your webserver:

```
<Location /aa-statistics>
	Options				+ExecCGI -Indexes
	AddHandler			cgi-script .go
	DirectoryIndex		clarin-sp-aagregator-golang.go

	AuthType            shibboleth
	ShibRequestSetting  requireSession 0	
	ShibUseHeaders      Off
	Satisfy             All
	Require             shibboleth
</Location>
```