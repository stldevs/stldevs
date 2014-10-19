select language, count(*) as count
from github.repo
group by language
order by count desc;
