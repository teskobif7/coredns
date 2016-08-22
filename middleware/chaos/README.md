# chaos

The `chaos` middleware allows CoreDNS to respond to TXT queries in the CH class.
This is useful for retrieving version or author information from the server.

## Syntax

~~~
chaos [version] [authors...]
~~~

* `version` the version to return. Defaults to CoreDNS-<version>, if not set.
* `authors` what authors to return. No default.

Note that you have to make sure that this middleware will get actual queries for the
following zones: `version.bind`, `version.server`, `authors.bind`, `hostname.bind` and
`id.server`.

## Examples

~~~
chaos CoreDNS-001 "Miek Gieben" miek@miek.nl
~~~
