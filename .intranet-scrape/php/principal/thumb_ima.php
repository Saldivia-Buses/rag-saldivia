<?php
// CREATE IMAGE
$url		= $_GET["url"];
$twidth 	= $_GET["ancho"];
$theight	= '';

if (isset($_GET["alto"]))
    $theight	= $_GET["alto"];
if (isset($_GET["Marca"]))
$MarcaAgua	= $_GET["Marca"];

$page		= (isset($_GET["page"])) ? $_GET["page"] : 0;

$dataPath = $_SESSION['datapath'];

if ($dataPath != ''){
    $tmpbase= '../database/'.$dataPath;
}
/*  tengo que obtener la extension */

$path_info		= pathinfo($url);
$extension		= $path_info["extension"];
$baseFileName	= $path_info["filename"];
$write = true;
//$extension = substr($url, strrpos($url, ".")+1);
$extension = strtolower($extension);

$maxalt = (isset($maxalt)) ? $maxalt : '';
$maxanch = (isset($maxanch)) ? $maxanch : '';
$tmphash = '';

$url = str_replace('xml/..','',$url);

//$tmphash = md5(realpath($url).'_'.$twidth.'_'.$theight.'_'.$maxalt.'_'.$maxanch);
if (is_file($url)){
    $tmphash = md5_file($url).'_'.$twidth.'_'.$theight.'_'.$maxalt.'_'.$maxanch;
}

include('../funciones/utiles.php');
$tmpFile = $tmpbase.'/tmp/'.$tmphash.'.jpg';
$display = (isset($display))?$display :'';

$rewrite = false;
//$rewrite = true;
if ($display != 'false')
if (!is_file($tmpFile) || $tmphash == '' || $rewrite == true){
	switch ($extension){
		case 'avi':
		case 'mov':
		case 'mpg':
		case 'mpeg':
		case 'wmv':
		case 'flv':
			$command = 'ffmpeg -y -i "'.$url.'"  -vcodec mjpeg -f rawvideo -ss 10 -vframes 1 -s 320x240 -an '.$tmpFile;
			exec($command);
			$image = new Imagick($tmpFile);
			 $image->setResolution( 300, 300 );
			$image->setImageFormat( "jpg" );
		break;
		case 'doc':
		case 'xls':
		case 'odt':
		case 'ods':
		case 'txt':
		case 'ppp':
			$image = convertFile($url, $page);
			if ($image){
			    $image->setResolution( 300, 300 );
			    $image->setImageFormat( "jpg" );
			}
		break;
		case 'pdf':
		    $myurl = $url.'['.$page.']';
		    $image = new Imagick();
		    
		    $image->setResolution( 300, 300 );
		    $image->readImage($myurl);                                    
			if(defined('Imagick::ALPHACHANNEL_OPAQUE')){
		  	  $image->setImageAlphaChannel(Imagick::ALPHACHANNEL_OPAQUE);
		  	}

		    $image->setImageFormat( "jpg" );

		break;
		case 'dwg';
			// CONVERTIR SVG A JPG
			$filesize  = @filesize	($url);
			$fixhash   = md5($url.$filesize);

			$tmpdwg = $tmpbase.'/tmp/'.$fixhash.'.dwg';
			$tmpsvg = $tmpbase.'/tmp/'.$fixhash.'.svg';

			$command = 'cp "'.$url.'" '.$tmpdwg;
			exec($command);

			if (!is_file($tmpsvg) ){
				$command = '../cgi-bin/cad2svg	"'.$tmpdwg.'" -o '.$tmpsvg.' 2>>'.$tmpbase.'/tmp/cad2svg.log';
				exec($command);
			}

			if (is_file($tmpsvg) ){
				$image = new Imagick($tmpsvg);
			}

			// FIX BLACK BACKGROUND and convert to JPG
			if (is_file($tmpsvg)){
				$command = '../cgi-bin/htx_svg_fix  "'.$tmpsvg.'" '.$tmpbase.'/tmp/tmp'.$tmphash.'.jpg ';
				exec($command);
			}

//			$command = '/usr/bin/convert  /tmp/'.$filename1.'.svg /tmp/'.$filename.'.jpg';
//			exec($command);


			//$url2='/tmp/tmp'.$tmphash.'.jpg';
			if (is_file($tmpsvg)){
				$image = new Imagick($tmpsvg);
			    $image->setResolution( 300, 300 );
				$image->setImageFormat( "jpg" );
			}
			else {
				$image = new Imagick();
				$pixel = new ImagickPixel( 'white' );

				/* New image */
				$image->newImage(100, 100, $pixel);

//				$image->setResolution( 300, 300 );
				$image->setImageFormat( "jpg" );
				$MarcaAgua = 'Sin imagen';
				$gravity = Imagick::GRAVITY_CENTER;
			}

		//	$image->negateImage(true);

		    //unlink($url2);
		break;
		case 'svg':                                          
		    $image = new Imagick($url);
		    $image->setResolution( 300, 300 );
		    $image->setImageFormat( "jpg" );
		break;
		default:
			$image = null;
			if (file_exists($url)){

				try{
			    $image = new Imagick($url);
			    } catch(Exception $ex){
			  	
			    }
			}

	}

	if ($image){

		if ($theight == '')
		    $theight = 0;

		if ($twidth == '')
		    $twidth = null;

		if ($twidth > 0 || $theight > 0) {

		//	$image->frameImage('white' ,0, 0, 0, 0); // hack to avoid black images

			// avoid imagick bug
			if ($theight == 0 || $twidth == 0){
			    $image->thumbnailImage($twidth,  $theight);
			} else {
			    $image->thumbnailImage($twidth,  $theight, true );
			}


			$alto = $image->getImageHeight();
			$ancho= $image->getImageWidth();
		}


		if ($write === true)
			$image->writeImage($tmpFile);
	}
	else die('error creating Image');


}else {

	$image = new Imagick($tmpFile);
}
if ($image != '' && $display != 'false') {
	header('Content-type: image/jpeg');

    $image->setImageFormat( "jpg" );

	if ($MarcaAgua) {
		if ($gravity =='')
			$gravity = Imagick::GRAVITY_CENTER;

		$ancho= $image->getImageWidth();
		$alto= $image->getImageHeight();

		/*
		$draw = new ImagickDraw();
		$draw->setGravity ($gravity);
		$pixel = new ImagickPixel( 'gray' );
		$draw->setFont('Bookman-DemiItalic');
		$draw->setFontSize( $ancho / 10 );
		// Create text
		$image->annotateImage($draw, 0, 0, 0, $MarcaAgua);
		*/
                            
		 $text	 = $MarcaAgua;
		 $font	 = 'Bookman-DemiItalic';
		 $font_size  = $ancho / 30;
		 $watermark  = array();

		 $draw = new ImagickDraw();
		 $draw->setGravity(Imagick::GRAVITY_CENTER);
		 $draw->setFont($font);
		 $draw->setFontSize($font_size);
		 $textColor = new ImagickPixel("white");
		 $draw->setFillColor($textColor);
		 $im = new imagick();
		 $properties = $im->queryFontMetrics($draw,$text);
		 $watermark['w'] = intval($properties["textWidth"] + 2);
		 $watermark['h'] = intval($properties["textHeight"] );
		 $im->newImage($watermark['w'],$watermark['h'],new ImagickPixel("transparent"));
		 $im->setImageFormat("jpg");
		 $im->annotateImage($draw, 0, 0, 0, $text);
		 $im->shadeImage(true, 70, 35);
//		 $im->blurImage(1,0.6);
		 //$im->normalizeImage();

		 $image->compositeImage( $im, Imagick::COMPOSITE_HARDLIGHT ,$ancho -  $watermark['w'], $alto - $watermark['h']);
	}
  
	echo $image;
}


function convertFile($url, $page){
	// CONVERTIR PDF A JPG
	$filesize  = @filesize	($url);
	$path_info	= pathinfo($url);
	$extension	= $path_info["extension"];
	$baseFileName	= $path_info["filename"];
	global $MarcaAgua;
	global $write;

	$filename = md5_file($url);

	if (!is_file($tmpbase.'/tmp/'.$filename.'.jpg')) {

		$tmppdf = $tmpbase.'/tmp/'.$filename.'.pdf';
		$uniqid = $tmpbase.'/tmp/'.$filename.'.'.$extension;

		if (!is_file($tmppdf)){
			copy($url, $uniqid);
			$convert = true;
				//$command = '../cgi-bin/oo2pdf '.$uniqid.' ';
				$command = 'abiword --to=pdf "'.$uniqid.'" ';

				//echo $command;
				$exec = exec($command);
		}

		$myurl = $tmppdf.'['.$page.']';
	    if (is_file($tmppdf)){
			$image = new Imagick($myurl);
			$image->setResolution( 300, 300 );
			$image->setImageFormat( "jpg" );
			//$image->writeImage('/tmp/'.$filename.'.jpg');
	    }
	}
	else {
		$image = new Imagick($tmpbase.'/tmp/'.$filename.'.jpg');
	}


   return $image;
}
?>
