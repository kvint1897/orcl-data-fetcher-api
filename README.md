# photo-fetcher

Web service for fetching photos from oracle DB

## setup as systemd service

```
sudo cp photo-fetcher-systemd.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl start photo-fetcher-systemd
sudo systemctl status photo-fetcher-systemd
sudo systemctl enable photo-fetcher-systemd
```
