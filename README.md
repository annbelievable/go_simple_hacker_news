# Simple Hacker News Clone

This Hacker news clone is implemented in golang. It will display the top 30 stories (excluding job posting). The whole flow works by getting an array of 500 top stories first, then try to get the top 40 stories with the provided id concurrently. The website will then filter out the job postings and execute the template and will only display the top 30 stories. For caching, I used go-cache which is a simple one. Both getting the item concurrently and also caching has speed up loading of the page a lot.

REFERENCES:
1) https://github.com/HackerNews/API
2) https://dev.to/sophiedebenedetto/synchronizing-go-routines-with-channels-and-waitgroups-3ke2
3) https://github.com/patrickmn/go-cache