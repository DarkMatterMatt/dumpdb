# The Plan

DumpDB imports credential dumps into a database to improve search performance.

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

- `databaseNames+`: One or more positional arguments of databases to initialise
- `conn`: connection string for the MySQL. Like user:pass@tcp(127.0.0.1:3306)
- `table="main"`: Table name for databases
- `sources=""`: Initialise the following database as the one to store sources in
- `sourcesTable="sources"`: Database table name to store sources in
- `engine="Aria"`: The database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL

## Process

Process files or folders into a regularised tab-delimited text file.

**Parameters:**

- `outputFileLines=4e6`: Number of lines per output file. 1e6 = ~32MB, 32e6 = ~1GB
- `outputFilePrefix="output_"`: Processed output file prefix
- `outputFileSuffix=".txt"`: Processed output file suffix
- `doneLog="[dbName]_done.log"`: Log file for processed input files
- `errLog="[dbName]_err.log"`: Log file for unparsed lines
- `skipLog="[dbName]_skip.log"`: Log file for skipped input files

### File Processing

- `.tar.gz`, `.tgz`: Decompress and open tarball, process each file
- `.txt`, `.csv`: Create `bufio.Scanner`
- `bufio.Scanner`: Process each line

## Import

Import files or folders into a database.

**Parameters:**

- `filesOrFolders+`: One or more positional arguments of files and/or folders to import
- `conn=`: Connection string for the SQL database. Like `user:pass@tcp(127.0.0.1:3306)/collection1`
- `table=`: Database table name to insert into
- `sourcesConn=`: Connection string for the sources database. Like `user:pass@tcp(127.0.0.1:3306)/sources`
- `sourcesTable="sources"`: Database table name to store sources in
- `engine="Aria"`: The database engine. Aria is recommended (requires MariaDB), MyISAM is supported for MySQL
- `compress`: Pack the database into a compressed, read-only format. Requires the Aria or MyISAM database engine
- `tmpFileLines=4e6`: Number of lines per temporary file (used for the LOAD FILE INTO command). 1e6 = ~32MB, 32e6 = ~1GB
- `tmpFilePrefix="[dbName]_"`: Temporary processed file prefix
- `tmpFileSuffix=".txt"`: Temporary processed file suffix
- `errLog="[dbName]_err.log"`: Log file for unparsed lines
- `doneLog="[dbName]_done.log"`: Log file for processed input files
- `skipLog="[dbName]_skip.log"`: Log file for skipped input files

## Search

Search multiple dump databases simultaneously.

**Parameters:**

- `query=""`: The WHERE clause of a SQL query. Yes it's injected, so try not to break your own database
- `columns="all"`: Comma separated list of columns to retrieve from the database
- `connPrefix=`: Connection string prefix to connect to MySQL databases. Like user:pass@tcp(127.0.0.1:3306)
- `db=`: Comma separated list of databases to search
- `dbTable="main"`: Database table name to search. Must be the same for all databases
- `sourcesConn=""`: Connection string for the sources database. Like `user:pass@tcp(127.0.0.1:3306)/sources`
- `sourcesTable="sources"`: Database table name to resolve integer sourceIDs to the source's name
