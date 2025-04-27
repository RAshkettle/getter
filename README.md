# getter

Getter is a faux Go web server.

## Usage

```
getter <folder_path>
```

Define a folder path and place any JSON files into the path to have it serve as a database. Each file will be treated as a table.

examples:

GET

localhost:9000/customers - returns a list of all customers in the file (just resturns the file contents as is)

localhost:9000/customers/5 = returns the record from the named file who's id matches the one provided.
