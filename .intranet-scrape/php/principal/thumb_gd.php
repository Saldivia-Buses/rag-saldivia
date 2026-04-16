<?php



$url 	 = $_GET["url"];
$twidth  = $_GET["ancho"];
$theight = $_GET["alto"];
$maxalt  = $_GET["maxalt"];
$maxanch = $_GET["maxanch"];
$MarcaAgua=$_GET["Marca"];
$docPreview=$_GET["docPreview"];

$page=$_GET["page"];
if ($page == '') $page= 0;
/* tengo que obtener la extension */

$path_info 		= pathinfo($url);
$extension 		= $path_info["extension"];
$baseFileName 	= $path_info["filename"];

//$extension = substr($url, strrpos($url, ".")+1);
$extension = strtolower($extension);

$tmphash = md5(realpath($url).'_'.$twidth.'_'.$theight.'_'.$maxalt.'_'.$maxanch);
$tmpFile = '/tmp/'.$tmphash.'.jpg';


if (!is_file($tmpFile)){
	if ($page == '') $page = 0;
    switch ($extension){
		case 'jpg':
		    $pic = @imagecreatefromjpeg($url) or die ("Image not found!");
		break;
		case 'png':
		    $pic = @imagecreatefrompng($url) or die ("Image not found!");
		break;
		case 'gif':
		  //  $pic = imagecreatefromgif($url) or die ("Image not found!");

			// CONVERTIR PDF A JPG
			$filename = uniqid('thumb');
			$command = '/usr/bin/convert  "'.$url.'[0]" /tmp/'.$filename.'.jpg';
			exec($command);
			$url2='/tmp/'.$filename.'.jpg';
		    $pic = @imagecreatefromjpeg($url2) or die ("Image not found!");
		    unlink($url2);


		break;
		case 'doc':
		case 'xls':
		case 'odt':
		case 'ods':
		case 'txt':
		case 'ppp':
           /* if ($docPreview == 'true'){
                $pic = convertFile($url, $page);
            } 
            else {
*/                $pic = @imagecreatefromjpeg('../img/mimetypes/misc.jpg');
  //          }
        break;

		case 'pdf':
			// CONVERTIR PDF A JPG
			$filename = uniqid('thumb');
			$command = '/usr/bin/convert "'.$url.'['.$page.']" /tmp/'.$filename.'.jpg';
			exec($command);
			$url2='/tmp/'.$filename.'.jpg';
		    $pic = @imagecreatefromjpeg($url2) or die ("Image not found!");
		   unlink($url2);
		break;
		case 'dwg':
			// CONVERTIR SVG A JPG

			$tmpdwg = '/tmp/'.$tmphash.'.dwg';
			$tmpsvg = '/tmp/'.$tmphash.'.svg';

			$command = 'cp "'.$url.'" '.$tmpdwg;
			exec($command);

			if (!is_file($tmpsvg)){
				$command = '../cgi-bin/cad2svg  "'.$tmpdwg.'" -o '.$tmpsvg.' 2>>/tmp/cad2svg.log';
				exec($command);
			}

			// FIX BLACK BACKGROUND and convert to JPG

			$command = '../cgi-bin/htx_svg_fix  "'.$tmpsvg.'" /tmp/tmp'.$tmphash.'.jpg ';
			exec($command);



//			$command = '/usr/bin/convert  /tmp/'.$filename1.'.svg /tmp/'.$filename.'.jpg';
//			exec($command);


			$url2='/tmp/tmp'.$tmphash.'.jpg';
		    $pic = @imagecreatefromjpeg($url2) or die ("Image not found!");
		    unlink($url2);
		break;
		case 'svg':
			// CONVERTIR SVG A JPG
			$filename = uniqid('thumb');
			$command = '/usr/bin/convert  "'.$url.'" /tmp/'.$filename.'.jpg';
			exec($command);
			$url2='/tmp/'.$filename.'.jpg';
		    $pic = @imagecreatefromjpeg($url2) or die ("Image not found!");
		    unlink($url2);
		break;
		default:
			$pic = @imagecreatefromjpeg('../img/mimetypes/misc.jpg');
		break;

    }

	header ("Content-type: image/jpeg"); # We will create an *.jpg

	if ($pic) {
	    $width = imagesx($pic);
	    $height = imagesy($pic);
		// Ancho por defecto
		if ($twidth=='')  $twidth = 225; # width of the thumb 160 pixel

		if ($theight == '') $theight = $twidth * $height / $width; # calculate height
		else $twidth = $theight * $width / $height; # calcula ancho

		/*if ($maxanch!=''){
			if ($width > $maxanch) {
				$twidth = $maxanch;
				$theight = $twidth * $height/ $width;
			}
		} */

		if ($maxalt!=''){
			if ($heigth > $maxalt) {
				$theigth = $maxalt;
				$twidth = $theight * $width / $height;
			}
		}


	    $thumb = imagecreatetruecolor ($twidth, $theight) or  die ("Can't create Image!");
		imagecopyresampled($thumb, $pic, 0, 0, 0, 0, $twidth, $theight, $width, $height); # resize image into thumb

		if ($MarcaAgua) {
			$font = 3;
			$txtColor = ImageColorAllocate($thumb, 0 , 0 ,0);
	//		$txtColor = ImageColorAllocate($thumb, 92 , 92 ,92);
	        	ImageString($thumb, $font, $twidth - ((strlen($MarcaAgua) + 1) * imagefontwidth($font)) ,
			    $theight - 1 * imagefontheight($font),  $MarcaAgua, $txtColor );

			$txtColor = ImageColorAllocate($thumb, 192 , 192 ,192);
	        	ImageString($thumb, $font, ($twidth - ((strlen($MarcaAgua) + 1) * imagefontwidth($font))) - 1 ,
			    ($theight - 1 * imagefontheight($font)) - 1,  $MarcaAgua, $txtColor );
		}

	    ImageJPEG($thumb,$tmpFile,75); # Thumbnail as JPEG
	    ImageDestroy($thumb);
	}
}

readfile($tmpFile);

function convertFile($url, $page){
			// CONVERTIR PDF A JPG
			$filesize  = @filesize  ($url);
			$filename = md5($url.$filesize);
			if (!is_file('/tmp/'.$filename.'.jpg')) {
				$tmppdf = '/tmp/'.$filename.'.pdf';
				$uniqid = '/tmp/'.$filename.'.doc';

				// require antiword aplication
				//$command = '/usr/bin/antiword -m 8859-1 -a a4 "'.$url.'" > /tmp/'.$filename.'.pdf';
				//exec($command);

				// oooconv converter (SLOW BUT BETTER)
				$command = 'cp "'.$url.'" '.$uniqid;
				exec($command);

				// TOO SLOW AND BUGGY


				$command = '../cgi-bin/oooconv "'.$uniqid.'"  '.$tmppdf. ' > /tmp/convlog.log';

//				$command = '../cgi-bin/oo2pdf '.$uniqid.' ';
				 exec($command);

				$command4 = '/usr/bin/convert "'.$tmppdf.'['.$page.']" /tmp/'.$filename.'.jpg';
				exec($command4);

				$url2='/tmp/'.$filename.'.jpg';
			}
			else
				$url2='/tmp/'.$filename.'.jpg';
		    $pic = @imagecreatefromjpeg($url2) or die ("Image not found!");
 		   	//unlink($url2);

		   return $pic;
}

?>