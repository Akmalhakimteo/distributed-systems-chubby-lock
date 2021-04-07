module ds_proj/m

go 1.16

replace ds_proj/server => ./server

require (
	ds_proj/client v0.0.0-00010101000000-000000000000 // indirect
	ds_proj/server v0.0.0-00010101000000-000000000000
<<<<<<< HEAD
	go.etcd.io/bbolt v1.3.5 // indirect
=======
	go.etcd.io/bbolt v1.3.5
>>>>>>> e7a029a8f831f4b5ad5ae7eee0f92698cd52f9cf
)

replace ds_proj/client => ./client
