Function ResizeImage {
	Param (
		[Parameter(Mandatory = $True, HelpMessage = "image in byte")]
		[ValidateNotNull()]
		$imageSrc,
		[Parameter(Mandatory = $true, HelpMessage = "Betwwen 16 and 1000")]
		[ValidateRange(16, 1000)]
		$canvasSize,
		[Parameter(Mandatory = $true, HelpMessage = "Between 1 and 100")]
		[ValidateRange(1, 100)]
		$ImgQuality,
		[Parameter(Mandatory = $True, HelpMessage = "Path for saving file")]
		[ValidateNotNull()]
		$Path
	)
	[void][System.Reflection.Assembly]::LoadWithPartialName("System.Drawing")
	$imageBytes = [byte[]]$imageSrc
	$ms = New-Object IO.MemoryStream($imageBytes, 0, $imageBytes.Length)
	$ms.Write($imageBytes, 0, $imageBytes.Length);
	$bmp = [System.Drawing.Image]::FromStream($ms, $true)
	# Image size after conversion
	$canvasWidth = $canvasSize
	$canvasHeight = $canvasSize
	# Set picture quality
	$myEncoder = [System.Drawing.Imaging.Encoder]::Quality
	$encoderParams = New-Object System.Drawing.Imaging.EncoderParameters(1)
	$encoderParams.Param[0] = New-Object System.Drawing.Imaging.EncoderParameter($myEncoder, $ImgQuality)
	# Get image type
	$myImageCodecInfo = [System.Drawing.Imaging.ImageCodecInfo]::GetImageEncoders() | Where-Object { $_.MimeType -eq 'image/jpeg' }
	# Get aspect ration
	$ratioX = $canvasWidth / $bmp.Width;
	$ratioY = $canvasHeight / $bmp.Height;
	$ratio = $ratioY
	if ($ratioX -ge $ratioY) {
		$ratio = $ratioX
	}
	# Create an empty picture
	$newWidth = [int] ($bmp.Width * $ratio)
	$newHeight = [int] ($bmp.Height * $ratio)
	$bmpResized = New-Object System.Drawing.Bitmap($newWidth, $newHeight)
	$graph = [System.Drawing.Graphics]::FromImage($bmpResized)
	$graph.DrawImage($bmp, 0, 0, $newWidth, $newHeight)
	# ----------------------------------------------------------------
	# Calculate the size of the square to crop
	$size = [Math]::Min($newWidth, $newHeight)	
	# Calculate the coordinates to crop the center of the image
	$x = ($newWidth - $size) / 2
	$y = ($newHeight - $size) / 2
	# Create a new bitmap with the square size
	$croppedImage = New-Object System.Drawing.Bitmap($size, $size)
	# Crop the image from the center
	$graph = [System.Drawing.Graphics]::FromImage($croppedImage)
	$graph.Clear([System.Drawing.Color]::White)
	$destRect = new-object System.Drawing.Rectangle(0, 0, $size, $size)
	$srcRect = new-object System.Drawing.Rectangle ($x, $y, $size, $size)
	$graph.DrawImage($bmpResized, $destRect, $srcRect, [System.Drawing.GraphicsUnit]::Pixel)
	$croppedImage.Save($Path, $myImageCodecInfo, $($encoderParams))
	# cleanup
	$croppedImage.Dispose()
	$bmpResized.Dispose()
	$bmp.Dispose()
}

$ADUserInfo = ([ADSISearcher]"(&(objectCategory=User)(SAMAccountName=$env:username))").FindOne().Properties
$ADUserInfo_sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
If ($ADUserInfo.pager) {
	$img_sizes = @(32, 40, 48, 96, 192, 200, 240, 448)
	$img_base = $env:public + "\AccountPictures"
	$reg_key = "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\AccountPicture\Users\$ADUserInfo_sid"
	$api_url = $args[0]
	$photo_url = "$api_url/photo/$($ADUserInfo.pager)"
	$photo = [byte[]](Invoke-WebRequest $photo_url).RawContentStream.ToArray()
	If ((Test-Path -Path $reg_key) -eq $false) {
		New-Item -Path $reg_key 
	}
	else {
		Write-Verbose "Reg key exist [$reg_key]" 
	}
	Try {
		ForEach ($size in $img_sizes) {
			$dir = $img_base + "\" + $ADUserInfo_sid
			If ((Test-Path -Path $dir) -eq $false) { $(New-Item -ItemType directory -Path $dir).Attributes = "Hidden" }
			$file_name = "Image$($size).jpg"
			$path = $dir + "\" + $file_name
			Write-Verbose " Crete file: [$file_name]"
			try {
				ResizeImage -imageSrc $photo -canvasSize $size -ImgQuality 100 -Path $path
				Write-Verbose " File saved: [$file_name]"
			}
			catch {
				If (Test-Path -Path $path) {
					Write-Warning "File exist [$path]"
				}
				else {
					Write-Warning "File not exist [$path]"
				}
			}
			$name = "Image$size"
			try {
				$null = New-ItemProperty -Path $reg_key -Name $name -Value $path -Force -ErrorAction Stop
			}
			catch {
				Write-Warning "Reg key edit error [$reg_key] [$name]"
			}
		}
	}
	Catch {
		Write-Error "Check permissions to files or registry."
	}
}