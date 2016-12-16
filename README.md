Drift
=====

## Migration SQL Specification
-

Revision File Grammar
```
revision     = {comments} section {section}
section      = {comments} changeset
changeset    = csheader {csheader} sql
csheader     = "--+" spaces "changeset" spaces "id:" runes  eol
id           = "id:" runes
pcheader     = "--+" spaces {rune} eol
pcsqlheader  = "--+" spaces {rune} eol
comments     = comment {comment}
comment      = "/*" spaces runes spaces "*/" |
               "//" spaces runes eol |
               "--" spaces runes eol
spaces       = space {space}
space        = " "
runes        = rune {rune}
rune         = ? int32 ?
eol          = "\n"
```

id:hello kitty author:jgilbert dbms:ql runalways:true, runonchange:true, failonerror:true
