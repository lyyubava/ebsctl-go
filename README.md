# ebsctl-go

ebsctl is a command-line tool designed to simplify managing Amazon Elastic Block Store (EBS) volumes. With ebsctl, users can create filesystems, mount volumes, and update the /etc/fstab file efficiently.

```
Usage:
  ebsctl [flags]

Flags:
      --dry-run             Do not run program, but show the list of actions the tool will perform
  -h, --help                help for ebsctl
      --label string        Filesystem label
      --mkfs string         Filesystem type to create
      --mountpoint string   Mountpoint for the volume
      --volume-id string    Volume id

```

## Example usage
```
go run main.go --mountpoint /mnt/data --label label --mkfs ext4 --volume-id vol-0940efdec80dbcddc  --dry-run true
```
