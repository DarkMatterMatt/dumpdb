# DumpDB

DumpDB imports credential dumps into a database to improve search performance.

There are two types of databases that will be created; one type stores the breach sources and the other type stores the dumped records. There should be a single sources-type database which stores where each record comes from (e.g. it could come from adobe2013 or collection1). There will be one or more databases which store the dumped records, these will be indexed and searched seperately.

## Installation

This project requires Go version 1.12 or later. You will also need access to a MariaDB (recommended) or MySQL server.

- `go get -u github.com/darkmattermatt/dumpdb`

## Example Usage

[Initialise](#init) the databases

```bash
go run github.com/darkmattermatt/dumpdb init -c "user:pass@tcp(127.0.0.1:3306)" -s sources -d adobe2013,collection1
```

[Import](#import) the dumped data

```bash
go run github.com/darkmattermatt/dumpdb import -c "user:pass@tcp(127.0.0.1:3306)" -s sources -d adobe2013 /path/to/data.tar.gz /more/data.txt
go run github.com/darkmattermatt/dumpdb import -c "user:pass@tcp(127.0.0.1:3306)" -s sources -d collection1 /path/to/data.tar.gz /more/data.txt
```

[Search](#search) the indexed data

```bash
go run github.com/darkmattermatt/dumpdb search -c "user:pass@tcp(127.0.0.1:3306)" -s sources -d adobe2013,collection1 -Q "email LIKE '%@example.com' LIMIT 10"
```

## General Info

**Verbosity:**

Output levels are as follows:

1. `RESULTS`: Only show errors and search results
1. `FATAL`: Only show errors and search results
1. `WARNINGS`: Nonfatal errors (usually occurring in one of the query threads)
1. `INFO`: The default level, provides minimal information at each step of the process
1. `VERBOSE`: Tells you what's going on
1. `DEBUG`: Spews out data

**Global Parameters:**

- `config=''`: Config file
- `v=3`: Verbosity. Set this flag multiple times for more verbosity
- `q=0`: Quiet. This is subtracted from the verbosity

## Init

Initialise a database for importing.

**Parameters:**

- `databases+`: One or more positional arguments of databases to initialise
- `databases=""`: Comma separated list of databases to initialise
- `conn=`: connection string for the MySQL. Like `user:pass@tcp(127.0.0.1:3306)`
- `sourcesDatabase=""`: Initialise the following database as the one to store sources in
- `engine="Aria"`: The database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL

## Process

Process files or folders into a regularised tab-delimited text file.

**Parameters:**

- `filesOrFolders+`: One or more positional arguments of files and/or folders to import
- `parser=`: The custom line parser to use. Modify the internal/parseline package to add another line parser
- `batchSize=4e6`: Number of lines per output file. 1e6 = ~64MB, 16e6 = ~1GB
- `filePrefix="[currentTime]_"`: Temporary processed file prefix

### File Processing

- `.tar.gz`, `.tgz`: Decompress and open tarball, process each file
- `.txt`, `.csv`: Create `bufio.Scanner`
- `bufio.Scanner`: Process each line

## Import

Import files or folders into a database.

**Parameters:**

- `filesOrFolders+`: One or more positional arguments of files and/or folders to import
- `parser=`: The custom line parser to use. Modify the internal/parseline package to add another line parser
- `conn=`: Connection string for the SQL database. Like `user:pass@tcp(127.0.0.1:3306)`
- `database=`: Database name to import into
- `sourcesDatabase=`: Database name to store sources in
- `compress=false`: Pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine
- `batchSize=4e6`: Number of results per temporary file (used for the LOAD FILE INTO command). 1e6 = ~64MB, 16e6 = ~1GB
- `filePrefix="[database]_"`: Temporary processed file prefix

**Notes:**

- By default, only the `mysql` user is able to read/write to the database file directly. A workaround is to run `go build .` and then `sudo -u mysql ./dumpdb import ...`

## Search

Search multiple dump databases simultaneously.

**Parameters:**

- `query=""`: The WHERE clause of a SQL query. Yes it's injected, so try not to break your own database
- `columns="all"`: Comma separated list of columns to retrieve from the database
- `conn=`: Connection string to connect to MySQL databases. Like `user:pass@tcp(127.0.0.1:3306)`
- `databases=`: Comma separated list of databases to search
- `sourcesDatabase=""`: Database name to resolve sourceIDs to their names from

**Notes:**

- The query is injected into the SQL command which means that any `LIMIT` statements are applied per database

## External Libraries

This project makes use of several excellent open-source libraries, listed below:

- [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)
- [github.com/mitchellh/go-homedir](https://github.com/mitchellh/go-homedir)
- [github.com/pbnjay/memory](https://github.com/pbnjay/memory)
- [github.com/spf13/cobra](https://github.com/spf13/cobra)
- [github.com/spf13/viper](https://github.com/spf13/viper)
