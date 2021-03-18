module ds_proj/m

go 1.16

replace ds_proj/server => ./server

require (
	ds_proj/client v0.0.0-00010101000000-000000000000 // indirect
	ds_proj/server v0.0.0-00010101000000-000000000000
)

replace ds_proj/client => ./client
