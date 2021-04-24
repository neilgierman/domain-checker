Write-Host "Both of these should return unknown"
(Invoke-WebRequest -Uri http://localhost:81/domains/example.com).Content
(Invoke-WebRequest -Uri http://localhost:81/domains/example2.com).Content
For ($i = 1; $i -lt 500; $i++) {
    Start-Job { Invoke-WebRequest -Method Put -Uri http://localhost:81/events/example.com/delivered; 
        Invoke-WebRequest -Method Put -Uri http://localhost:82/events/example.com/delivered; 
        Invoke-WebRequest -Method Put -Uri http://localhost:82/events/example2.com/delivered; 
        Invoke-WebRequest -Method Put -Uri http://localhost:81/events/example2.com/delivered }
}
Get-Job | Wait-Job | Out-Null
Write-Host "Both of these should still return unknown because we wrote 998 deliveries to example.com and example2.com"
(Invoke-WebRequest -Uri http://localhost:81/domains/example.com).Content
(Invoke-WebRequest -Uri http://localhost:81/domains/example2.com).Content
Invoke-WebRequest -Method Put -Uri http://localhost:81/events/example.com/delivered | Out-Null
Invoke-WebRequest -Method Put -Uri http://localhost:82/events/example.com/delivered | Out-Null
Write-Host "Now the first should be catch-all and the second still unknown"
(Invoke-WebRequest -Uri http://localhost:81/domains/example.com).Content
(Invoke-WebRequest -Uri http://localhost:81/domains/example2.com).Content
Invoke-WebRequest -Method Put -Uri http://localhost:81/events/example.com/bounced | Out-Null
Write-Host "Now the first should be not catch-all and the second still unknown"
(Invoke-WebRequest -Uri http://localhost:81/domains/example.com).Content
(Invoke-WebRequest -Uri http://localhost:81/domains/example2.com).Content
