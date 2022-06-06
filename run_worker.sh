for IDX in {1..20}
do
  go run ./worker > "./log/worker_{$i}.log" &
done