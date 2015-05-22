# seec: See Commmit

## Goal

- Extract committer and author information about a commit from git/git
- Produces a markdown text ready to be copy-pasted from clipboard

## Usage

    seec <sha1>

## Example

```
seec 05c39674f35f33b6d2311da6c63268b9e7739840

See [commit ed178ef](https://github.com/git/git/commit/ed178ef13a26136d86ff4e33bb7b1afb5033f908) by [Jeff King](https://github.com/peff (`peff`)), 22
Apr 2015.  
<sup>(Merged by [Junio C Hamano](https://github.com/gitster -- `gitster` --) in [commit 05c3967](https://github.com/git/git/commit/05c39674f35f33b6d2311da6
c63268b9e7739840), 19 May 2015)</sup>
```

## Notes

- Uses Go-GitHub librarie: https://github.com/google/go-github
- Uses clipboard for golang: https://github.com/atotto/clipboard
- Made for Stack Overflow contributions like http://stackoverflow.com/a/30375581/6309
