### Notes
- The example has a sample exchange of an index chain.
- It stores the links in a datastore (we can probably use in destination a temporary memory datastore so we don't
have to keep a dedicated datastore in the indexer for ingestion.

### TO DO
- Instead of fetching the full IPLD DAG, use stop condition to stop at the link
that we already have. 
        - Reference: https://github.com/ipld/go-ipld-prime/commit/e44329e855d8b9a8643c7b00b97bdbcdc1496c03
- Upgrade go-legs and this example to the latest linkSytem: https://github.com/ipld/go-ipld-prime/blob/master/linkingExamples_test.go
- Clean code.
