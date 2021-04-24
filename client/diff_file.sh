#!/bin/sh
echo "Running bash script";
end=$(($1-1))
for i in $(seq 1 $end)
do
	val=$(($1*$2+$i))
	new_file="$3$val"
	echo $new_file
	app $val $new_file &
done
val=$(($1*$2))
new_file="$3$val"
app $val $new_file