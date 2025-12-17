@echo off
set FILTER_BRANCH_SQUELCH_WARNING=1
"C:\Program Files\Git\bin\bash.exe" -c "git filter-branch -f --tree-filter 'if [ -f DEVELOPMENT_LOG.md ]; then sed -i.bak \"s/AIzaSyDFxR0BZrJbMgPseUVDbsoa9VJc2rOOvxw/REMOVED_API_KEY/g\" DEVELOPMENT_LOG.md && rm -f DEVELOPMENT_LOG.md.bak; fi' --prune-empty --tag-name-filter cat -- --all"

