# ebsctl-go

ebsctl is a command-line tool designed to simplify managing Amazon Elastic Block Store (EBS) volumes. With ebsctl, users can create filesystems, mount volumes, and update the /etc/fstab file efficiently.
```
go run main.go --mountpoint /mnt/data --label label --mkfs ext4 --volume-id vol-0940efdec80dbcddc  --dry-run true
```
