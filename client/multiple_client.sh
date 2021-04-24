#!/bin/sh
echo "Running bash script";
end=$(($1-1))
for i in $(seq 1 $end)
do
	val=$(($1*$2+$i))
	app $val $3 &
done
val=$(($1*$2))
app $val $3