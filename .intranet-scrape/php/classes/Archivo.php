<?php
/*
 * Created on 04/01/2007
 */

class Archivo {
 	/* Array de Datos*/
    var $nombre;
    var $descripcion;
    var $extension;
    var $tipo;
    var $size;
    var $base;
    var $path;

    var $filesize;
    var $mimetype;

    var $icono;
    var $html;
    var $link;

    // para imagenes
    var $width;
    var $height;


    public function __construct($nombre, $path='', $base='../archivos/', &$ObjCampo=null,  &$Datos=null) {
        $this->nombre = $nombre;
        $this->path = $path;
        $this->base = $base;
        $this->creolink();
	    $this->uid = uniqid();
        $this->determinoTipo();
        $this->ObjCampo = $ObjCampo;
        $this->Datos = $Datos;
    }

    function creolink() {
        if ($this->path != '' &&  substr($this->path, -1 , 1) != '/') $this->path .='/';

        $this->link = $this->base.$this->path.$this->nombre;

    }

    function showIcono() {
        $id= '_img_'.$this->ObjCampo->NombreCampo;
        if (isset($this->imgId))
                    $id = $this->imgId;
        if ($this->icono != '')
            return '<img  id="'.$id.'" border="0" src="../img/mimetypes/'.$this->icono.'" />';
    }

    function imageLink($width='', $height='', $MarcaAgua='', $params=''){
            $urllink = urlencode($this->link);
            $Imagen   = 'thumb.php?url=' . $urllink .'&ancho='.$width.'&alto='.$height.'&Marca='.$MarcaAgua.$params;
	    return $Imagen     ;
    }
    
    function thumb($width=22, $height=22, $MarcaAgua='', $params='', $imgParams='') {

        if ($this->imagen == true || $this->preview ) {

            //$urllink = urlencode($this->link);            
            // 'thumb.php?url=' . $urllink .'&ancho='.$width.'&alto='.$height.'&Marca='.$MarcaAgua.$params;

            $Imagen   = $this->imageLink( $width, $height, $MarcaAgua, $params);
            $miImagen = '<img border="0" src="' . $Imagen . '" '.$imgParams.'/>';
        }

        if ($this->mimetype=='image/svg') {
            $miImagen = '<embed src="'.$this->link.'" width="'.$width.'" height="'.$height.'" type="image/svg+xml" />';
        }

        if ($this->mimetype == 'audio/mpeg' || $this->tipo == 'flv') {
            $miImagen = $this->mediaPlayer($this->link);
        }

        if ($miImagen == '') {
            $icon = '../img/mimetypes/'.$this->icono;
            $Imagen   = 'thumb.php?url=' . $icon .'&ancho='.$width.'&alto='.$height.'&Marca='.$MarcaAgua;
            $miImagen = '<img border="0" src="' . $Imagen . '" '.$imgParams.'/>';
        //	$miImagen .=  $this->mimetype.'<br>';

        }

        return $miImagen;
    }

/**
* HTML5 media player
*/
//    function html5Player($file, $type="audio/wave") 
    function html5Player($file, $type="audio/wave") 
    {
    	if (!is_file($file))
	       return;

//      $downlink = $file;
//      $downlink = Archivo::downloadLink($file);
    
    	$tag = '<audio controls="true" title="'.basename($file).'" preload="none" height="20px" '.$type.'>';
    	$tag .= 'audio not supported';
    	$tag .= '<source src="'.$file.'" />';
        $tag .= '</audio>';

        return $tag;

    }


    function mediaPlayer($file) {
    // VER para hacer streaming
    //  $streamer = $this->downloadLink($file, 'true');
        $width = 120;
        $height = 20  ;
        if ($this->tipo == 'flv') {
            $height = 100  ;
        }
        $idvid = uniqid('video');
        $flv = '<span id="'.$idvid.'"></span>';
        $flv .='<script type="text/javascript">';
        $flv .='var s1 = new SWFObject("mediaplayer.swf","mediaplayer","'.$width.'","'.$height .'","8");';
        $flv .='s1.addParam("allowfullscreen","true");';

        $flv .='s1.addVariable("width","'.$width.'");';
        $flv .='s1.addVariable("backcolor","0x000000");';
        $flv .='s1.addVariable("frontcolor","0xFFFFFF");';
        $flv .='s1.addVariable("lightcolor","0xFF8A23");';
        $flv .='s1.addVariable("showdigits","true");';
        $flv .='s1.addVariable("file","'.$file.'");';

        $flv .='s1.addVariable("type","'.$this->tipo.'");';
        if ($this->tipo == 'flv') {
            $flv .='s1.addVariable("height","'.$height.'");';
            $Imagen = substr($file, 0, strpos( $file, '.flv')).'.jpg';
            $flv .='s1.addVariable("image","'.$Imagen.'");';
        }

        $flv .='s1.write("'.$idvid.'");';
        $flv .='</script>';
        return $flv;

    }

    /**
     * SWF Document Viewer
     */
    function swfViewer($file) {
        $width = '100%';
        $height ='94%' ;
        $idvid = uniqid('video');
        $flv = '<span id="'.$idvid.'"></span>';
        $flv .='<script type="text/javascript">';
        $flv .='var s1 = new SWFObject("'.$file.'","mediaplayer","'.$width.'","'.$height .'","8");';
        $flv .='s1.addVariable("height","'.$height.'");';
        $flv .='s1.addVariable("width","'.$width.'");';
        $flv .='s1.write("'.$idvid.'");';
        $flv .='</script>';
        return $flv;
    }

    function showInline($orden='', $alternatePath='') {
        $ObjCampo = $this->ObjCampo;
        $this->lightbox = true;
        $valor = $this->nombre;
        $nonutflink = $this->link;
        $this->link = $this->link;
        $file = rawurlencode($this->link);
        //$file = $this->link;
        $link = $ObjCampo->url.$file;
        
        $genPath =$this->Datos->path;
        if ($ObjCampo->url == '') {
            $filePath= $genPath;
        }
        
        // Add Object Path
        if ($ObjCampo->path !='' && $ObjCampo->tipoDato != 'dir' && $this->Datos != null) {
        
            $tempField = $this->Datos->getCampo($ObjCampo->path);
            if (is_object($tempField)){
                $Objurl = (isset($tempField->valor))?$tempField->valor:'';
    
            }
            else { 
                $Objurl = $ObjCampo->path;
            }
                
            if ($orden !== '') {
                $Tablatemp = $this->Datos->TablaTemporal->datos();
                $Objurl = (isset($Tablatemp[$orden][$ObjCampo->path]))? $Tablatemp[$orden][$ObjCampo->path]:$ObjCampo->path ;
            }
            if ($Objurl == '') $Objurl = $alternatePath;
            $filePath .= '/'.$Objurl;
        }

        $link  = '../database/'.$_SESSION['datapath'].'xml/'.$filePath.'/'.$link;
        $link2 = '../database/'.$_SESSION['datapath'].'xml/'.$filePath.$nonutflink;

        if(is_dir($link2)) {
            $btnlink = new Html_button(basename($valor), '../img/folder.png');
            $btnlink->addParameter('title', $valor);
            $hash= md5($valor);
            $basename= basename($link2);
            $access= $ObjCampo->access;
            $urlencodePath = urlencode($genPath.$nonutflink);
             $click= "xmlLoader('$hash', '&url=fileManager.php&amp;basedir=/$urlencodePath&access=$access', {title:'$basename', loader:'cargahttp.php'});";
            $this->lightbox = false;

            $btnlink->addEvent('onclick', $click);

            $valor = $btnlink->show();
        }
        /* Per type  */
        switch ($this->tipo) {
            case "flv":
            case "mp3":
                $valor = $this->mediaPlayer($link);
	        $this->lightbox = false;
                
                break;
            case "ogg":
                $valor = $this->html5Player($link, 'audio/ogg');
	        $this->lightbox = false;
                return $valor;
                break;
            case "wav":
                $valor = $this->html5Player($link, 'audio/wave');
	        $this->lightbox = false;
                return $valor;
                break;

        }


        if ($this->ObjCampo->download == "true" || $this->download == true) {
            $imgdown= '<img  align="center"  src="thumb.php?url='.$link.'&amp;ancho=100" height="26px" />';
            //$imgdown= '<img src="../img/go-down.png" />';

            $downloadLink = $this->downloadLink($link);
            $nombre= basename($link);
            $onclick  = 'onClick="window.open(\''.$downloadLink.'\');';
            $uid= uniqid('img');
            $valor = '<button type="button" onMouseOver="showImage(this, \''.$link.'\', event, null, \''.$nombre.'\', \''.$uid.'\');"  onMouseOut="cerrarVent(\''.$uid.'\');" title="'.$this->nombre.'" '.$onclick.'">'.$imgdown.'</button>';
            $this->lightbox = false;

	    return $valor;            
            //die($valor);
        }


        if ($this->ObjCampo->open == "true" || $this->open == true) {
            $ancho=($ObjCampo->Size != '')?$ObjCampo->Size:80;

            $image = 'thumb.php?url='.$link.'&amp;ancho='.$ancho;

            if ($ObjCampo->height != ''){
                $height= $ObjCampo->height;
                $image .= '&alto='.$height;
		
	    }


            $btnOpen = new Html_button($label, $image  );
            $nombre= urldecode(basename($link));
            $btnOpen->addParameter('name', 'Open');
            $downloadLink = $this->downloadLink($link, '', '&').'true';
            $uid= uniqid('img');
            
            $btnOpen->addEvent('onMouseOut', 'cerrarVent(\''.$uid.'\')' );
            $btnOpen->addEvent('onMouseOver', 'showImage(this, \''.$link.'\', event, null, \''.$nombre.'\', \''.$uid.'\');' );
            $btnOpen->addEvent('onclick', 'Histrix.loadInnerXML(\''.$this->Datos->xml.'\', \'cargahttp.php?url='.$downloadLink.'&stream=true\', null,  \''.$nombre.'\' , null, null, {maximize:true});');
            $valor = $btnOpen->show();

            $this->lightbox = false;

	    return $valor;
            //die($valor);
        }


        if ($this->ObjCampo->viewer == "true" || $this->viewer == true) {
            $ancho=($ObjCampo->Size != '')?$ObjCampo->Size:80;

            $image = 'thumb.php?url='.$link.'&amp;ancho='.$ancho;

            if ($ObjCampo->height != ''){
                $height= $ObjCampo->height;
                $image .= '&alto='.$height;
		
	    }


            $btnOpen = new Html_button($label, $image  );
            $nombre= urldecode(basename($link));
            $btnOpen->addParameter('name', 'Open');
            //$downloadLink = $this->downloadLink($link, '', '&').'true';
            $downloadLink = $link;
            $uid= uniqid('img');
            
            $path = dirname($link);

            $file = basename($link);
            $title= $file;
            $file = urlencode  ( $file  );
	    $sessionvar = uniqid('prev');
            $_SESSION[$sessionvar]= $path;

            $downlink =  'pdfviewer.php?ro=true&dir='.$sessionvar.'&f='.$file.'&ancho=1024&alto=768';            
            $btnOpen->addEvent('onMouseOut', 'cerrarVent(\''.$uid.'\')' );
            $btnOpen->addEvent('onMouseOver', 'showImage(this, \''.$link.'\', event, null, \''.$nombre.'\', \''.$uid.'\');' );
            $btnOpen->addEvent('onclick', 'Histrix.loadInnerXML(\''.$this->Datos->xml.'\', \''.$downlink.'\', null,  \''.$nombre.'\' , null, null, {width:\'80%\',height:\'98%\', modal :true});');
	
            $valor = $btnOpen->show();

            $this->lightbox = false;

	    return $valor;
	    
            //die($valor);
        }



        // JUST IMAGES
        if (isset($this->imagen) && $this->imagen == true) {
            if (isset($this->svg) && $this->svg ==true) {
                $label = 'Ver';
                $btnSVG = new Html_button($label, "../img/mimetypes/image.png" ,$label );
                $btnSVG->addParameter('name', 'SVG');
                $btnSVG->addEvent('onclick', 'Histrix.loadInnerXML(\''.$valor.'\', \'svgrender.php?url='.$link.'\', \'\',\''.$this->nombre.'\', \'_'.$this->Datos->xml.'\');');
                $btnSVG->tabindex = $this->tabindex();
                $btn = $btnSVG->show();
                $valor = $btn.'<embed src="'.$link.'" width="100" height="50" type="image/svg+xml" />';
            }
            else {
                $ancho=($ObjCampo->Size != '')?$ObjCampo->Size:80;
                $nombre= urldecode(basename($link));
                //	$valor = '<img  onMouseOut="cerrarVent(\''.$link.'\');" onMouseOver="showImage(this, \''.$link.'\', event, null, \''.$nombre.'\');" src="thumb.php?url='.$link.'&amp;ancho='.$ancho.'"  />';
                //	$valor = '<img alt="'.$link.'"onMouseOut="cerrarVent(\''.$link.'\');" onMouseOver="showImage(this, \''.$link.'\', event, null, \''.$nombre.'\');" src="thumb.php?url='.$link.'&amp;ancho='.$ancho.'"  />';

                $uid= uniqid('img');
                
                $valor = '<img class="inlineImg" _alt="'.$link.'"onMouseOut="cerrarVent(\''.$uid.'\');" onMouseOver="showImage(this, \''.$link.'\', event, null, \''.$nombre.'\', \''.$uid.'\');" src="thumb.php?url='.$link.'&amp;ancho='.$ancho.'"/>';

            }
        }
        
        if ($this->lightbox == true){
                $valor = '<a href="thumb.php?url='.$link.'&alto=600" rel="lightbox['.$this->uid.']" >'.$valor.'</a>';
        
        }
        
        //$valor .= $link2;
        return $valor;
    }

    function viewerButton() {
        $path = dirname($this->link);
        $file = basename($this->link);

        $title= utf8_decode($file);
        $title= $file;

        $file = urlencode  ( $file  );
        $sessionvar = uniqid('prev');
        $_SESSION[$sessionvar]= $path;
        $downlink =  'pdfviewer.php?dir='.$sessionvar.'&f='.$file.'&ancho=650&alto=500&tipo='.$this->tipo;
        $onclick = 'Histrix.loadExternalXML (\'pdf\', \''.$downlink.'\');';

        $onclick = "Histrix.loadInnerXML ('$sessionvar', '$downlink', null, '$title',  null, null,  {width:'80%', height:'90%', modal:true})";

        $viewer = new Html_button('', '../img/revision16.png');
        $viewer->addParameter('title', 'Preview');
        $viewer->addEvent('onclick', $onclick);


        return $viewer;

    }

    function downloadLink($link='', $stream='' , $inline = '?') {
        if ($link == '') {
            $link = $this->link;
        }
        $path = dirname($link);
        $file = basename($link);
        $sessionvar = uniqid('down');
        $_SESSION[$sessionvar]= $path;
        $downlink =  'download.php'.$inline.'dir='.$sessionvar.'&f='.$file.'&stream='.$stream.'&DAT='.$_SESSION['DAT'];
        return $downlink;
    }

    function downloadButton($link='', $file='', $image='') {
        $downlink = Archivo::downloadLink($link);

        $download = '<a  title="descargar"  class="boton" style="vertical-align: top;display:inline-block;padding-right:3px;padding-left:3px; height:auto;" target="_blank" href="'.$downlink.'">';
        if ($image != '')
            $download .= $image;
        else
            $download .= '<img style="border:0px;" src="../img/go-down.png" />';
        $download .= $file;
        $download .= '</a>';
        return $download;
    }

    function printButton($link='', $file='') {
        if ($link == '') {
            $link = $this->link;
        }
        $downlink = $link;
        $print = '<a  title="'.$this->i18n['print'].'" class="boton" style="vertical-align: top;display:inline-block;padding-right:3px;padding-left:3px; height:28px;" target="print" href="'.$downlink.'">';
        $print .= '<img style="border:0px;" src="../img/printer1.png" />';
        $print .= $file;
        $print .= '</a>';
        return $print;
    }

    function determinoTipo() {


        $this->extension = substr($this->nombre, strrpos($this->nombre, ".")+1);
        if (is_file($this->link)) {
            $path_info 		= pathinfo($this->link);
            $extension 		= $path_info["extension"];
           // $baseFileName 	= $path_info["filename"];
            $this->extension = $extension;
            $this->filesize  = @filesize  ($this->link );
            $this->mimetype  = @mime_content_type  ( $this->link );
        }
        $ext = strtolower($this->extension);
        $icon = 'unknown.png';
        switch ($ext) {
            case "odt" :
            case "doc" :
                $this->tipo = $ext;
                $icon = 'wordprocessing.png';
                $this->preview = true;
            break;
            case "xls" :
            case "ods" :

                $this->tipo = $ext;
                $icon = 'spreadsheet.png';
                $this->preview = true;
            break;
            case "txt" :
                $this->tipo = $ext;
                $icon = 'txt.png';
                $this->preview = true;
            break;
            case "pdf" :
                $this->tipo = $ext;
                $this->imagen = true;
                $icon = 'pdf.png';
                $this->preview = true;
            break;
            case "zip" :
                $this->tipo = $ext;
                $icon = 'tgz.png';
            break;
            case "swf" :
                $this->tipo = $ext;
                $icon = 'flash.png';
                $this->preview = true;
            break;
            case "mp3" :
            case "wav" :
            case "ogg" :
                $this->imagen = false;
                $this->html5Audio = true;
                $this->tipo = $ext;
                $icon = 'sound.png';
            break;
            case "avi" :
            case "mov" :
            case "mpg" :
            case "mpeg" :
            case "wmv" :
            case "flv" :

                $this->tipo = $ext;
                $this->preview = true;

                //$icon = 'video.png';
                break;
            case "dwg" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;
                $this->preview = true;

                break;
            case "svg" :
            case "eps" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;
            break;
            case "tiff" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;
//                $myImg = @imagecreatefromjpeg($this->link);
                break;
            
            case "jpg" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;
                $myImg = @imagecreatefromjpeg($this->link);
                break;
            case "jpeg" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;
                $myImg = @imagecreatefromjpeg($this->link);
                break;
            case "gif" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;

                $myImg = @imagecreatefromgif($this->link);
                break;
            case "png" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $myImg = @imagecreatefrompng($this->link);
                $this->imagen = true;
                break;
            case "bmp" :
                $this->tipo = $ext;
                $icon = 'image.png';
                $this->imagen = true;
                break;
        }
        $this->icono = $icon;

        if (isset($this->imagen) && $this->imagen) {
            $this->exif = @exif_read_data (  $this->link, 0, true );
            $this->preview = true;

            if ($myImg) {
                $this->width = imagesx($myImg);
                $this->height  = imagesy($myImg);
            }

        }

    }

    // image representacion of File
    // requires Imagick
    //
    function toImage($url, $twidth=0, $theight=0, $page=0, $dpi=300) {
    
        $dataPath = $_SESSION['datapath'];

        if ($dataPath != ''){
            $tmpbase= '../database/'.$dataPath;

        }

        /*  tengo que obtener la extension */

        $path_info 	= pathinfo($url);
        $extension 	= $path_info["extension"];
        $baseFileName 	= $path_info["filename"];
        $write = true;
        //$extension = substr($url, strrpos($url, ".")+1);
        $extension = strtolower($extension);

	if ($this->ObjCampo->imageCrop != ''){
	    $crop = explode(',', $this->ObjCampo->imageCrop);

	}

        $tmphash = md5(realpath($url).'_'.$twidth.'_'.$theight.'_'.$page.'_'.$dpi.$this->ObjCampo->imageCrop);
        $tmpFile = $tmpbase.'/tmp/'.$tmphash.'.jpg';
//        $refresh = true;
        if (!is_file($tmpFile) || $refresh == true) {

            switch ($extension) {
                case 'avi':
                case 'mov':
                case 'mpg':
                case 'mpeg':
                case 'wmv':
                case 'flv':
                    $command = 'ffmpeg -y -i "'.$url.'"  -vcodec mjpeg -f rawvideo -ss 10 -vframes 1 -s 320x240 -an '.$tmpFile;
                    exec($command);
                    $image = new Imagick($tmpFile);
                    $image->setResolution( $dpi, $dpi );
                    $image->setImageFormat( "jpg" );
                    break;
                case 'doc':
                case 'xls':
                case 'odt':
                case 'ods':
                case 'txt':
                case 'ppp':
                    $image = convertFile($url, $page);
                    if ($image) {
                        $image->setResolution( $dpi, $dpi );
                        $image->setImageFormat( "jpg" );
                    }
                    break;
                case 'pdf':
                    $myurl = $url.'['.$page.']';
                   
                    $image = new Imagick($myurl);
                    $image->setResolution( $dpi, $dpi );
		    $image->setImageAlphaChannel(Imagick::ALPHACHANNEL_OPAQUE);
                    
                    $image->setImageFormat( "jpg" );
                    break;
                case 'dwg';
                // CONVERTIR SVG A JPG
                    $filesize  = @filesize  ($url);
                    $fixhash   = md5($url.$filesize);

                    $tmpdwg = $tmpbase.'/tmp/'.$fixhash.'.dwg';
                    $tmpsvg = $tmpbase.'/tmp/'.$fixhash.'.svg';

                    $command = 'cp "'.$url.'" '.$tmpdwg;
                    exec($command);

                    if (!is_file($tmpsvg) ) {
                        $command = '../cgi-bin/cad2svg  "'.$tmpdwg.'" -o '.$tmpsvg.' 2>>'.$tmpbase.'/tmp/cad2svg.log';
                        exec($command);
                    }

                    if (is_file($tmpsvg) ) {
                        $image = new Imagick($tmpsvg);
                    }

                    // FIX BLACK BACKGROUND and convert to JPG
                    if (is_file($tmpsvg)) {
                        $command = '../cgi-bin/htx_svg_fix  "'.$tmpsvg.'" '.$tmpbase.'/tmp/tmp'.$tmphash.'.jpg ';
                        exec($command);
                    }

                    //			$command = '/usr/bin/convert  /tmp/'.$filename1.'.svg /tmp/'.$filename.'.jpg';
                    //			exec($command);


                    //$url2='/tmp/tmp'.$tmphash.'.jpg';
                    if (is_file($tmpsvg)) {
                        $image = new Imagick($tmpsvg);
                        $image->setResolution( $dpi, $dpi );
                        $image->setImageFormat( "jpg" );
                    }
                    else {
                        $image = new Imagick();
                        $pixel = new ImagickPixel( 'white' );

				/* New image */
                        $image->newImage($dpi, $dpi, $pixel);

                        //				$image->setResolution( 300, 300 );
                        $image->setImageFormat( "jpg" );
                        //$MarcaAgua = 'Sin imagen';
                        //$gravity = Imagick::GRAVITY_CENTER;
                    }

                    //	$image->negateImage(true);

                    //unlink($url2);
                    break;
                case 'svg':
                    $image = new Imagick($url);
                    $image->setResolution( $dpi, $dpi );
                    $image->setImageFormat( "jpg" );
                    break;
                default:
		    try{
                	$image = new Imagick($url);
                	$image->setResolution( $dpi, $dpi );
                	$image->setImageFormat( "jpg" );
		    } catch (Exception $e){
			loger($url, 'image_error.log');
			// image not found
		    }

            }
            if ($image) {

		if ($crop != ''){
	    	    $image->cropImage( $crop[0],$crop[1],$crop[2],$crop[3]);
		}



                if ($twidth > 0 ) {
                    $image->thumbnailImage($twidth, 0);
                    $alto= $image->getImageHeight();

                    if ($theight != '' && $alto > $theight) {
                        $image->thumbnailImage(0, $theight);
                    }
                }

                if ($theight > 0 ) {
                    $image->thumbnailImage(0, $theight);
                    $ancho= $image->getImageWidth();

                    if ($twidth != '' && $ancho > $twidth) {
                        $image->thumbnailImage($twidth, 0);
                    }
                }
                if ($this->watermark != ''){
                     $text   = $this->watermark;
                     //$font   = 'Bookman-DemiItalic';
                     $font   = 'Helvetica-Bold';

                     $ancho= $image->getImageWidth();
                     $font_size  = $ancho / 30;
                     $watermark  = array();

                     $draw = new ImagickDraw();
                     $draw->setGravity(Imagick::GRAVITY_CENTER);
                     $draw->setFont($font);
                     $draw->setFontSize($font_size);
                     $textColor = new ImagickPixel("black");
                     $strokeColor = new ImagickPixel("white");
                     $draw->setFillColor($textColor);
                     $draw->setStrokeWidth  ( 2 );
                     $draw->setStrokeColor($strokeColor);
                     $im = new imagick();
                     $properties = $im->queryFontMetrics($draw,$text);
                     $watermark['w'] = intval($properties["textWidth"] + 2);
                     $watermark['h'] = intval($properties["textHeight"] );
                     $im->newImage($watermark['w'],$watermark['h'],new ImagickPixel("transparent"));
                     
                     
                     $im->setImageFormat("jpg");
                     $im->annotateImage($draw, 0, 0, 0, $text);
                     //$im->shadeImage(true, 70, 35);

                     $image->compositeImage( $im, Imagick::COMPOSITE_HARDLIGHT ,$ancho -  $watermark['w'], $alto - $watermark['h']);
                }




                if ($write === true)
                    $image->writeImage($tmpFile);

            }
//            else die('error creating Image');

        }else {
            $image = new Imagick($tmpFile);
        }

	if ($image){
	    $this->width  = $image->getImageWidth();
            $this->height = $image->getImageHeight();
	}
        
        return $tmpFile;
    }
    

    function byteConvert($bytes) {
        $s = array('B', 'Kb', 'MB', 'GB', 'TB', 'PB');
        $e = floor(log($bytes)/log(1024));

        return @sprintf('%.2f '.$s[$e], ($bytes/pow(1024, floor($e))));
    }
                    /* EXIF INFO
	foreach ($exif as $key => $section) {
		if ($key == 'IFD0')
	    foreach ($section as $name => $val) {
	    	if ($name == 'UndefinedTag:0xC4A5') continue;
	    	if (trim($val) == '') continue;
	        $foto .= "$name: $val".'<br>';
	    }
	}
	*/
}
 ?>
