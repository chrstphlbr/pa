# *pa* - Performance (Change) Analysis

*pa* is an efficient tool for analysing performance data based on statistical simulation, i.e., bootstrap.
The performance data can come from measurements such as performance tests or benchmarks.

There are two main ways how to analyse performance data with *pa*:
1. Single version analysis: statistic (e.g., mean) + variability (confidence interval of the statistic)
2. Analysis between two versions: confidence interval ratio of a statistic (e.g., mean)

Inspired by [1], *pa* employs a Monte-Carlo technique called *bootstrap* [2] to estimate the population confidence interval from a given sample.
It uses hierarchical random re-sampling with replacement [3].
In the context of performance analysis, these hierarchical levels correspond to levels where measurement repetition happens to have reliable results.
These hierarchical levels are:
1. invocation
2. iteration
3. fork
4. trial
5. instance

Higher levels are composed of multiple occurences of lower levels, i.e., an iteration (level 2) consists of many invocations (level 1), and so on.

## Requirements and Installations

*pa* requires only Go with version 1.13 or higher ([Install Page](https://golang.org/doc/install)).

Install *pa* by running `go get github.com/chrstphlbr/pa`.



## Usage


### Command Line Interface

*pa* comes with a simple command line interface (optional flags in `[...]` with their defaults):

```bash
pa  [-bs 10000] [-is 0] [-sl 0.01] [-st mean] [-os] [-m 1] \
    file_1 \
    [file_2 ... file_n] 
```

If 1 file (`file_1`) is provided, the single version analysis is performed, i.e., the confidence intervals of a single performance experiment is computed.

If multiple files are provided, the two version analysis is performed:
the confidence intervals for both versions *and* the confidence interval ratio between the two versions is computed.
In the simple case, 2 files are provided, `file_1` for version 1 and `file_2` for version 2.
It is also possible to provide multiple files per version (of equal number) by setting the flag `-m`. 

Note that the files **MUST** be sorted alphabetically by their benchmarks (see section "Input Files").

Flags:
* `-bs` defines the number of bootstrap simulations, i.e., how many random samples are taken to estimate the population distribution
* `-is` defines how many invocation samples are used (0 takes the mean across all invocations of an iteration, -1 takes all invocations, and > 0 for number of samples).
* `-sl` defines the significance level.
The confidence level for the confidence intervals is then `1-sl`.
The default is 0.01 which corresponds to a 99% confidence level.
* `-st` defines the statistic for which a confidence interval is computed.
The deafult is `mean`.
Another option is `median`. 
* `-os` defines whether the statistic, as set by `-st`, is included in the output file.
* `-m` sets the number of files per version (test and control group).
For example, if `-m 3` *pa* expects 6 files, where `file_1`, `file_2`, and `file_3` belong to version 1, and `file_4`, `file_5`, and `file_6` belong to version two.


### Input Files

*pa* expects CSV input files of the following form.
For JMH benchmark results, the tool [bencher](https://github.com/chrstphlbr/bencher) can transform JMH JSON output to this CSV file format.
```
project;commit;benchmark;params;instance;trial;fork;iteration;mode;unit;value_count;value
```

The columns represent the following values:
1. `project` is the project name
2. `commit` is the project version, e.g., a commit hash
3. `benchmark` is the name of the fully-qualified benchmark method
4. `params` are the performance parameters (not the function/method parameters) of the benchmark in comma-separated form.
Every parameter consists of a name and a value, separated by an equal sign (`name=value`).
For example JMH supports performance parameters through its `@Param` annotation
5. `instance` is the name of the instance or machine (level 5)
6. `trial` is the number of the trial (level 4)
7. `fork` is the fork number (level 3).
For example JMH supports forks through their `@Fork` annotation
8. `iteration` is the iteration number within a fork (level 2)
9. `mode` is the benchmark mode.
For exmaple JMH supports average time `avgt`, throughput `thrpt`, or sample time `sample`
10. `unit` is the measurement unit of the benchmark value.
Depending on the `mode`, the measurement unit can be ns/op for average time or op/s for throughput 
11. `value_count` is the number of invocations (level ) the `value` occurred in this iteration.
Every iteration can have multiple values (i.e., invocations), which are presented as a histogram.
Each histogram value corresponds to one CSV row, and the occurrences of this value is defined by `value_count`.
12. `value` is the performance metric with a certain `unit`

**IMPORTANT**: the input files must be sorted by `benchmark` and `params`, otherwise the tool will not work correctly.
This is because input files can be *large* and, therefore, *pa* works on file input streams.


### Output File

Output files can contain 3 types of rows:
* rows starting with `#` are comments
* empty rows
* all other rows are CSV rows

The 
1. `benchmark` is the name of the benchmark
2. `params` are the function/method parameters of the benchmark.
*pa* does not populate this column, because the input format does not provide the function/method parameters
3. `perf_params` is a comma-separated list of performance parameters.
See column `params` of the input files for comparison
4. `st` is the statistic the confidence interval is for.
Can be "mean" or "median"
5. `ci_l` is the lower bound of the confidence interval
6. `ci_u` is the upper bound of the confidence interval
7. `cl` is the confidence level of the confidence interval

#### Single Version Analysis

The output file is a CSV with the following columns (without `-os`):
```
benchmark;params;perf_params;ci_l;ci_u;cl
```

And with the statistic, as set by `-os`, it has the following columns:
```
benchmark;params;perf_params;st;ci_lower;ci_u;cl
```

#### Two Version Analysis

The output file is a CSV with the following columns (without `-os`):
```
benchmark;params;perf_params;v1_ci_l;v1_ci_u;v1_cl;v2_ci_l;v2_ci_u;v2_cl;ratio_s;ratio_ci_l;ratio_ci_u;ratio_cl
```

And with the statistic, as set by `-os`, it has the following columns:
```
benchmark;params;perf_params;v1_st;v1_ci_l;v1_ci_u;v1_cl;v2_st;v2_ci_l;v2_ci_u;v2_cl;ratio_st;ratio_ci_l;ratio_ci_u;ratio_cl
```

Compared to the single version analysis, the two version analysis has three or four (with or without `-os`) columns, for both versions (`v1` and `v2`) and the confidence interval for the ratio between the two versions (`ratio`).



## References

[1] T. Kalibera and R. Jones, “Quantifying performance changes with effect size confidence intervals,” University of Kent, Technical Report 4–12, June 2012. Available: [URL](http://www.cs.kent.ac.uk/pubs/2012/3233)

[1] A. C. Davison and D. V. Hinkley, “Bootstrap methods and their application”

[2] S. Ren, H. Lai, W. Tong, M. Aminzadeh, X. Hou, and S. Lai, “Nonparametric bootstrapping for hierarchical data”, Journal of Applied Statistics, vol. 37, no. 9, pp. 1487–1498, 2010. Available: [DOI](https://doi.org/10.1080/02664760903046102)
