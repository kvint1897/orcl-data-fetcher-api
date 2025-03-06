# orcl-data-fetcher-api

Web service for fetching data from oracle DB

## setup as systemd service

```
sudo cp orcl-data-fetcher-api-systemd.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl start orcl-data-fetcher-api-systemd
sudo systemctl status orcl-data-fetcher-api-systemd
sudo systemctl enable orcl-data-fetcher-api-systemd
```
