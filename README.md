# CHia Log sTats

CLI tool to easily parse Chia log files and understand phase durations at a glance.

## Output

```
May 5, 2021
---------------------------------------------------------------------------------------------------------
K     RAM     Threads    Phase 1    Phase 2    Phase 3    Phase 4    Copy       Total      Start    End
32    3390    3:65536    4h 20m     1h 28m     2h 57m     0h 10m     1h 45m     10h 40m    21:45    08:25
32    3390    3:65536    4h 21m     1h 27m     2h 57m     0h 10m     1h 47m     10h 42m    21:45    08:27
32    3390    3:65536    4h 23m     1h 27m     2h 56m     0h 10m     1h 47m     10h 43m    21:45    08:28
32    3390    3:65536    4h 24m     1h 26m     2h 56m     0h 10m     1h 47m     10h 43m    21:45    08:28

May 6, 2021
--------------------------------------------------------------------------------------------------------
K     RAM     Threads    Phase 1    Phase 2    Phase 3    Phase 4    Copy       Total      Start    End
32    3390    2:65536    2h 37m     2h 58m     5h 13m     0h 12m     0h 28m     11h 28m    22:00    09:27
32    3390    2:65536    3h 34m     2h 52m     4h 22m     0h 9m      0h 29m     11h 27m    23:00    10:26
32    3390    2:65536    4h 41m     2h 25m     3h 31m     0h 8m      0h 9m      10h 54m    00:00    10:54

May 7, 2021
--------------------------------------------------------------------------------------------------------
K     RAM     Threads    Phase 1    Phase 2    Phase 3    Phase 4    Copy       Total      Start    End
32    3390    2:65536    3h 31m     1h 51m     3h 2m      0h 10m     0h 18m     8h 52m     23:59    08:51
32    3390    2:65536    3h 56m     1h 40m     2h 51m     0h 9m      0h 28m     9h 4m      00:24    09:28
32    3390    2:65536    4h 3m      1h 41m     2h 39m     0h 8m      0h 24m     8h 55m     00:49    09:44

Config Averages
-------------------------------------------------------------------------------------------------
K     RAM      Threads    Phase 1    Phase 2    Phase 3    Phase 4    Copy       Total      Plots
32    3390     3:65536    4h 7m      1h 28m     2h 58m     0h 10m     1h 26m     10h 9m     11
32    3814     3:65536    3h 39m     1h 18m     2h 33m     0h 10m     1h 9m      8h 49m     3
32    3400     2:65536    2h 56m     1h 8m      2h 20m     0h 9m      0h 37m     7h 11m     2
32    3400     3:65536    4h 44m     1h 23m     2h 53m     0h 10m     1h 10m     10h 20m    3
32    3390     2:65536    3h 36m     1h 38m     2h 55m     0h 11m     0h 47m     9h 6m      26

Parallel Averages
-----------------------------------------------------------------------------------
Phase 1    Phase 2    Phase 3    Phase 4    Copy       Total      Parallel    Plots
4h 10m     1h 52m     3h 54m     0h 16m     0h 33m     10h 44m    1           3
3h 38m     1h 32m     2h 47m     0h 10m     0h 49m     8h 55m     3           24
4h 16m     1h 30m     2h 56m     0h 10m     1h 22m     10h 14m    4           16
2h 56m     1h 8m      2h 20m     0h 9m      0h 37m     7h 11m     2           2
```

## Options

### `-f` output format

Sets output format. Can be piped to file.

`chlots -f csv > file.csv`

**Supported options**
* default *default*
* csv

### `-a` display averages

Displays config and parallel averages

`chlots -a`
