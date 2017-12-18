# clarin-sp-aagregator-golang
This project is a cgi-scipt, implemented in golang, with the aim to collect and send shibbolleth attribute information to the CLARIN aagregator.

The implementation is a port of the php based [clarin-sp-aaggregator](https://github.com/ufal/clarin-sp-aaggregator).

# Configuration

## Shibboleth Service Provider
Follow the guidelines provided [here](https://github.com/ufal/clarin-sp-aaggregator#shibboleth2xml). Bottom line is to add `exportLocation` and `exportACL` attributes to your applications `session` definition in `shibboleth2.xml`. See [NativeSPSessions](https://wiki.shibboleth.net/confluence/display/SHIB2/NativeSPSessions) for more information.

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

The cgi executable can be customized via the following parameters:

| Name            | Default Value                    | Description                                 |
| --------------- | -------------------------------- | ------------------------------------------- |
| submit_sp_stats | false                            | If true submit stats to remote service      |
| log_path        | /var/log/sp-session-hook/        | Directory where the log file will be stored | 
| log_file        | session-hook.log                 | Name of the log file                        |
| aag_url         | https://clarin-aa.ms.mff.cuni.cz | URL of the remote endpoint                  |
| aag_path        | /aaggreg/v1/got                  | Path of the service                         |
| sp_entity_id    | https://sp.catalog.clarin.eu     | Specify hardcoded entity id                 |

Use the `SetEnv` directive in apache to set these variables, see [mod_env](http://httpd.apache.org/docs/current/mod/mod_env.html) and [env](http://httpd.apache.org/docs/current/env.html) for more info.

```
SetEnv submit_sp_stats false
```
