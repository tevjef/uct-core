#! /bin/bash

go build github.com/tevjef/uct-core/scrapers/cuny
go build github.com/tevjef/uct-core/common/tools/uct-clean

declare -a arr=("BAR"
                "BMC"
                "BCC"
                "BKL"
                "LAW"
                "MED"
                "SPH"
                "CTY"
                "CSI"
                "NCC"
                "HOS"
                "HTR"
                "JJC"
                "KCC"
                "LAG"
                "LEH"
                "MEC"
                "NYT"
                "QNS"
                "QCC"
                "SPS"
                "GRD"
                "YRK")
for i in "${arr[@]}"
do
    time ./cuny -c ../../../common/conf/config.toml -u "$i" -f json > "$i"-D.json
    ./uct-clean -f json -o json "$i"-D.json >  "$i".json
    rm "$i"-D.json
done