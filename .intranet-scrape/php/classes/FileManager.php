<?php
/*
 * File Manager class 2009-07-06
 *
 */

class FileManager {
    var $modo;
    var $inputField;
    var $dirvar;
    
    public function __construct($slashdir, $basePath, $access) {
        $this->slashdir = $slashdir;
        $this->basePath = $basePath;
        $this->access   = $access;
    }

    // Navigation Bar
    public function navbar($direc, $inputField, $basedir, $dirvar, $midir, $midir2 , $enableBack) {
        $this->currentDir = $basedir;
                
        $salida .= '<div style="float:left;" >';
        $salida .= $this->backButton( $inputField, $basedir, $dirvar, $midir, $enableBack);
        $salida .= $this->homeButton($inputField, $basedir, $dirvar);
        $salida .= $this->linkButtons($direc, $inputField, $basedir, $dirvar, $midir2);
        $salida .= '</div>';
        return $salida;
    }

    public function statusBar() {

    }
    /** link buttons
     *
     * @param array $direc
     * @param string $inputField input field to return value to
     * @param string $basedir 
     * @param string $dirvar
     * @param string $midir2
     * @return string  buttons to quick navigation
     */
    public function linkButtons($direc, $inputField, $basedir, $dirvar,$midir2) {
        foreach($direc as $nfile1 => $fileant) {
            if ($fileant == '') continue;

            $midir2 .= $fileant.'/';
            if ($fileant == '..') continue;

            $link = new Html_button($fileant);
            $link->addParameter('style', 'padding-top:3px;padding-bottom:4px;',true);
            $location = 'window.location=\'?inputField='.$inputField.'&basedir='.$basedir.'&dirvar='.$dirvar.'&dir2='.urlencode($midir2).'&DAT='.$_SESSION['DAT'].'\'';
            $link->addEvent('onclick', $location);
            $salida .= $link->show();
        }
        return $salida;
    }

    public function homeButton( $inputField, $basedir, $dirvar) {
        $link = new Html_button('','../img/go-home.png');
        $link->addEvent('onclick', 'window.location=\'?inputField='.$inputField.'&dirvar='.$dirvar.'&basedir='.$basedir.'&DAT='.$_SESSION['DAT'].'\'');
        $salida = $link->show();
        return $salida;
    }

    public function backButton( $inputField, $basedir, $dirvar, $midir, $enableBack) {
    // Back Button
        $atras = new Html_button('', '../img/go-previous.png');
        $location = 'window.location=\'?inputField='.$inputField.'&dirvar='.$dirvar.'&basedir='.$basedir.'&dir2='.$midir.'&DAT='.$_SESSION['DAT'].'\'';
        $atras->addEvent('onclick', $location);
        if (!$enableBack) $atras->addParameter('disabled', 'disabled');
        $salida = $atras->show();
        return $salida;
    }

    public function uploadButton($inputField, $dirvar, $basedir, $slashdir, $modo) {

        $salida .= '<div style="float:left;">';
        $subir = new Html_button('', '../img/upload.png');
        $subir->addParameter('title', 'Subir Archivo');
        $subir->addEvent('onclick', 'toggleID(\'uploadFile\');');
        $salida .= $subir->show();
        $salida .= '<div class="boton" style="border:1px solid #ccc; display:none; float:right;" id="uploadFile" >';


        $salida .= '<form style="display:inline;" id="uploadForm" name="uploadForm" method="POST"  enctype="multipart/form-data" action="?dirvar='.$dirvar.'&inputField='.$inputField.'&basedir='.$basedir.'&dir2='.$slashdir.'&DAT='.$_SESSION['DAT'].'"> ';
        $salida .=  '<input type="hidden" name="MAX_FILE_SIZE" value="90000000" />';
        $salida .=  '<input type="hidden" name="modo" value="'.$modo.'" />';
        $salida .=  '<input type="hidden" name="maxWidth" value="'.$this->maxWidth.'" />';


     //   $salida .= '<span  id="output"></span>';
        $salida .= '<div id="divStatus"></div>';
        /*
        $subir2 = new Html_button('Elegir Archivos', '../img/uploadMultiple.png');
        $subir2->addParameter('title', 'Subir Multiples Archivo');
        $subir2->addParameter('id', 'buttonPlaceHolder');
        $subir2->addEvent('onclick', 'swfu.selectFiles();');
        $salida .= $subir2->show();
*/

        $send = new Html_textBox('Elegir', '');
        $send->addParameter('title', 'Crear');
        $send->addParameter('name', 'filename');
        $send->addParameter('type', 'file');
        $salida .= $send->show();


        $ok = new Html_button('', '../img/ok.png');
        $ok->addParameter('title', 'Subir');
        $ok->addParameter('type', 'submit');
        $ok->addEvent('onclick', 'submit');
        $salida .= $ok->show();

        $cancel = new Html_button('', '../img/cancel.png');
        $cancel->addParameter('title', 'Cancelar');
        $cancel->addParameter('id', 'btnCancel');
        $cancel->addEvent('onclick', 'swfu.cancelQueue(); toggleID(\'uploadFile\');');
        $salida .= $cancel->show();
        

        $salida .= '</form>';
        $salida .= '</div>';
        $salida .= '</div>';

        return $salida;
    }
    public function uploadJavascript(){
        $basePath = $this->basePath;
        $basedir  = $this->basedir;
        $slashdir = $this->slashdir;

        return '<script type="text/javascript">


        </script>';
        
    }
    
    public function refreshButton ($inputField, $basedir, $dirvar, $slashdir) {
        $refresh = new Html_button('', '../img/view-refresh.png');
        $refresh->addParameter('title', 'Refrescar');
        $refresh->addParameter('id', 'RefreshFileManager');
        $refresh->addEvent('onclick', 'window.location=\'?inputField='.$inputField.'&basedir='.$basedir.'&dirvar='.$dirvar.'&dir2='.urlencode($slashdir).'&DAT='.$_SESSION['DAT'].'\'');
        $salida .= $refresh->show();
        return $salida;
    }

    public function viewAsButton($inputField, $basedir, $dirvar, $slashdir, $modo) {
        $opciones['lista'] = 'ver como lista';
        $opciones['icons'] = 'ver como iconos';
        $selectModo =  new Html_select($opciones);
        $selectModo->value 	= $modo;
        $selectModo->addEvent('onchange', 'window.location=\'?inputField='.$inputField.'&dirvar='.$dirvar.'&basedir='.$basedir.'&modo=\'+this.value+\'&dir2='.urlencode($slashdir).'&DAT='.$_SESSION['DAT'].'\'');
        $salida .= $selectModo->show();
        return $salida;
    }

    public function mkdirButton ($inputField, $dirvar, $basedir, $slashdir) {

    // FOLDER CREATION

        $folder = new Html_button('', '../img/folder-new.png');
        $folder->addParameter('title', 'Nueva Carpeta');
        $folder->addEvent('onclick', 'toggleID(\'newfolder\');');

        $salida .= '<div style="float:right;" >';
        $salida .= $folder->show();


        $salida .= '<span _class="boton" style="padding:10px; display:none;" id="newfolder" >';
        $urlaction = '?dirvar='.$dirvar.'&folder=true&inputField='.$inputField.'&basedir='.$basedir.'&dir2='.$slashdir.'&DAT='.$_SESSION['DAT'];
        $salida .= '<form style="display:inline;" id="createFolder" name="createfolder" method="post" action="'.$urlaction.'"> ';
        $salida .=  '<input type="hidden" name="modo" value="'.$modo.'" />';



        $input = new Html_textBox('', 'varchar');
        $input->addParameter('name', 'folder');
        $salida .= $input->show();

        $ok = new Html_button('', '../img/ok.png');
        $ok->addParameter('title', 'Crear');
        $ok->addParameter('type', 'submit');
        $ok->addEvent('onclick', 'submit');
        $salida .= $ok->show();

        $cancel = new Html_button('', '../img/cancel.png');
        $cancel->addParameter('title', 'Cancelar');
        $cancel->addEvent('onclick', 'toggleID(\'newfolder\');');

        $salida .= $cancel->show();

        $salida .= '</form>';
        $salida .= '</span>';
        $salida .= '</div>';
        return $salida;
    }

    public function toolbar($inputField, $dirvar, $basedir, $slashdir, $modo, $basePath, $access, $mkfolder, $folderName) {

        $salida .= '<div style="width:auto; display:inline;float:right; clear:both;position:absolute;right:10px;">';
       // $access = $this->access;
        if (strpos($access, 'w') !== false) {

            $salida .= $this->uploadButton($inputField, $dirvar, $basedir, $slashdir, $modo);
            $salida .= $this->mkdirButton ($inputField, $dirvar, $basedir, $slashdir);

            if($mkfolder == 'true') {
                $newdir = utf8_encode($basePath.$basedir.$slashdir.$folderName);
                if (!is_dir($newdir))
                	mkdir($newdir, 0777, true);
            }

        }

        $salida .= $this->refreshButton($inputField, $basedir, $dirvar, $slashdir);
        $salida .= $this->viewAsButton( $inputField, $basedir, $dirvar, $slashdir, $modo);

        $salida .= '</div>';

        return $salida;

    }
    public function message( $message='') {
        if ($message == '') $this->message;
        $salida .= '<div id="output" _id="fsUploadProgress" style="padding-right:1em; padding-left:1em ;z-index:1000; position:relative;float:left;" >';
        $salida .= $mensaje;
        $salida .= '</div>';
        $salida .= '<div id="currentDir" style="display:none;">'.$this->currentDir.$this->slashdir.'</div>';
        $salida .= '<table id="speedStatus" ></table>';
        return $salida;
    }

    public function topBar(  $midir, $midir2, $access,  $FILES) {
        $slashdir = $this->slashdir;
        $basePath = $this->basePath;
        $basedir  = $this->basedir;

        $inputField = $this->inputField;
       // $access   = $this->access;
       $dirvar = $this->dirvar;
        $salida  .= '<div class="barraUrl" style=" top:0px; height:30px;">';


        $direc = explode('/', $slashdir);
        $cantdirs= count($direc);
        foreach($direc as $nfile1 => $fileant) {

            if ($fileant == '') continue;
            if ($nfile1 < $cantdirs - 2)
                $midir .= $fileant.'/';
            $enableBack=true;

            // DISABLE HIDDEN DIRECTORIES
            if (substr($fileant, 0,1)=='.') {

                $access = 'rd';
            }
        }


        $mensaje='';
        if ($FILES['filename']['tmp_name'] != '') {

            $nombre = basename($FILES['filename']['name']);
            $uploadfile = $basePath.$basedir.$slashdir.$nombre;

            $save_path = $basePath.$basedir.$slashdir;
            // Validate that we won't over-write an existing file
            // VERSIONING

            if (file_exists($uploadfile)) {
                    //HandleError("File with this name already exists");
                    //exit(0);

                    $md5Old= md5_file($uploadfile);
                    $md5New= md5_file($FILES['filename']['tmp_name']);

                    //VERSIONING FILE SYSTEM
                    // make hidden folder
                    if ($md5Old != $md5New){
                            $path_info 		= pathinfo($uploadfile);
                            $extension 		= $path_info["extension"];
                            $baseFileName 	= $path_info["filename"];

                            $hiddenFolder = $save_path.'.'.$nombre;
                            $filetime = date('Y_m_d_H:i:s',filemtime  ( $save_path . $nombre  ));
                            $newFileName  = $baseFileName.'_'.$filetime.'.'.$extension;
                            @mkdir($hiddenFolder);
                            rename($save_path . $nombre, $hiddenFolder.'/'. $newFileName);
                    }
                    else {
                       // echo 'son iguales '.$md5Old.' -->'.$md5New;
                    }

            }




            if (move_uploaded_file($FILES['filename']['tmp_name'], $uploadfile)) {
                $mensaje = 'El Archivo: <b>'.$nombre. '</b> ha sido recibido exitosamente.';
                chmod( $uploadfile, 0666);

                // resize Image if needeed
                $maxwidth = $_REQUEST['maxWidth'];
		if ($maxwidth != '' && $maxwidth > 0){
		    try {
		        $image = new Imagick($uploadfile);
			if ($image->getImageWidth() > $maxwidth){
		    //    $image->setResolution( 300, 300 );
			    $image->adaptiveResizeImage($maxwidth, 0);
	        	    $image->writeImage($uploadfile);

		    	}

		    } catch(Exception $ex) {
			// not a valid Image
		    }
		}
                
                
                
            }
            else {
                $mensaje = 'Error al Subir el Archivo: '.  $FILES['filename']['error'];
            }
        }




        $salida .= $this->navbar($direc, $inputField, $basedir, $dirvar, $midir, $midir2 , $enableBack);
        $salida .= $this->message($mensaje);
        $newFolder = $_POST['folder'];
        
        $salida .= $this->toolbar($inputField, $dirvar, $basedir, $slashdir, $modo, $basePath, $access, $_GET['folder'], $newFolder);
        $salida .= '</div>';
        return $salida;
    }


    function getDirContents( $basedir2, $slashdir='', $cant =0 ) {

        //global $fileManager;

        $modo       = $this->modo;
        $inputField = $this->inputField;
        $dirvar     = $this->dirvar;
        $access     = $this->access;
        $basedir    = $this->basedir;



        if (substr($basedir2, -1 ,1) != '/')  $basedir2.= '/'; // add trail slash
        if (substr($slashdir,  1 ,1) == '/')  $slashdir = substr($slashdir, 1 ); // remove trail slash

   //     $basePath  = str_replace('//','/',$basePath);
        $basedir   = str_replace('//','/',$basedir);
        $basedir2  = str_replace('//','/',$basedir2);
        $slashdir  = str_replace('//','/',$slashdir);
        
        if (!is_dir  ($basedir2.$slashdir )) {

            $mkdir= @mkdir($basedir2.$slashdir, 0777, true);
            
            if (!$mkdir) {
                echo 'Unable to create: '.$basedir2.$slashdir.' check permisions';
                //errorMsg('Unable to create: '.$basedir2.$slashdir.' check permisions');
                return false;
            }
        }
        if(is_dir($basedir2.$slashdir)) {
            $dh = opendir($basedir2.$slashdir);
        }
        else {
            errorMsg('Unable to open: '.$basedir2.$slashdir);
            return false;
           // die();
        }
        $arrayimg='';


        $files = scandir ($basedir2.$slashdir);
        $midir = '';
        foreach($files as $nfile => $file) {

            if (is_dir($basedir2.$slashdir . $file) && $file != "."   && $file != ".." ) {

            //$descripcion= utf8_decode($file);
                $descripcion= $file;
                $midir = $basedir2.$slashdir.$file;
                if ($file=='.') continue;
                if ($file=='..') continue;
                if ( substr($file,0,1 ) == '.' )continue;


                $thumb = '';
                $deletebutton = '';
                $stat =  stat( $basedir2.$slashdir.$file );
                $ultimaMod = $stat['mtime'];

                $contenido = $this->getDirContents( $basedir2, $slashdir.$file, 1);
                $hay = false;
                $cantidad = 0;
                if ($contenido)
                    foreach($contenido as $n => $confile) {
                        $cantidad ++;
                        if ($hay === true) continue;

                        if (strpos(strtolower($confile), ".jpg"))  $extension = 'jpg';
                        if (strpos(strtolower($confile), ".jpeg")) $extension = 'jpg';
                        if (strpos(strtolower($confile), ".tiff"))  $extension = 'tiff';
                        if (strpos(strtolower($confile), ".gif")) $extension = 'gif';
                        if (strpos(strtolower($confile), ".png")) $extension = 'png';
                        if (strpos(strtolower($confile), ".pdf")) $extension = 'pdf';
                        if (strpos(strtolower($confile), ".doc")) $extension = 'doc';
                        if (strpos(strtolower($confile), ".svg")) $extension = 'svg';
                        if (strpos(strtolower($confile), ".dwg")) $extension = 'dwg';

                        if ($extension!= '') {
                            $hay = true;
                            $thumb = $this->miniThumb($basedir2.$slashdir.$file.'/'.$confile, $docPreview);
                        }
                    }
                else {
                    $hash = md5($basedir2.$slashdir .$file);

                    if (strpos($access, 'w') !== false) {
                        if ($_GET['del'] != '' && $_GET['del'] == $hash) {
                            rmdir($basedir2.$slashdir . $file);
                            $_GET['del'] = '';
                            continue;
                        }


                        $deletedir = new Html_button('', '../img/cancel.png');
                        $deletedir->addParameter('style', 'float:right; bottom:2px; right:2px');
                        $deletedir->addParameter('title', 'Borrar');
                        $deletedir->addEvent('onclick', "fmDelete('$slashdir', '$hash', '$basedir', '$dirvar', '".$_SESSION['DAT']."');");
                        $deletebutton = $deletedir->show();
                    }

                }

                switch ($modo) {
                    case 'icons':

                        $href 	 = '<a href="?basedir='.$basedir.'&dirvar='.$dirvar.'&dir2='.urlencode($slashdir.$file).'/'.'&DAT='.$_SESSION['DAT'].'">';
                        $dirspan = '<span class="dir" style="height:50px;width:100px">';
                        $titulo	 = '<span class="dirname">   '.$descripcion.'   </span>';

                        if ($deletebutton != '' ) {
                            $dir  = $dirspan;
                            $dir .= $href;
                            $dir .= $titulo;
                            $dir .= '</a>';
                            $dir .= $deletebutton;
                            $dir .= '</span>';

                        }
                        else {
                            $dir = $href;
                            $dir .= $dirspan;
                            $dir .= $titulo;
                            $dir .= $thumb;
                            $dir .= '</span>';
                            $dir .= '</a>';
                        }



                        break;
                    case 'lista':
                        $href = '<a href="?basedir='.$basedir.'&dirvar='.$dirvar.'&dir2='.urlencode($slashdir.$file).'/'.'&DAT='.$_SESSION['DAT'].'">';
                        $folder = '<img border="0" src="../img/folder.png" />';

                        $dir = '<tr>';
                        $dir .= '<td>'.$href.$folder.$descripcion.'</a>'.'</td>';
                        $dir .= '<td>'.$cantidad. ' elementos'.'</td>';
                        $dir .= '<td>'.$deletebutton.'</td>';
                        $dir .= '<td>Carpeta</td>';
                        $dir .= '<td>'.date('d/m/Y - H:i:s',$ultimaMod).'</td>';

                        $dir .= '</tr>';

                        break;
                }

                if ($cant == 0)
                    $salida .= $dir;

            } else
            if ($file != "." && $file != "..") {
                $i++;
                $arrayimg[$i] = $file;
            }

        }
        if ($cant != 0) return $arrayimg;

        if ($arrayimg != '') {
            $salida .=  $this->showGal($basedir2.$slashdir, $arrayimg, $slashdir, $modo);
        }

        closedir($dh);

        return $salida;
    }


    function showGal( $path, $imagenes,  $slashdir, $modo='lista' ) {

        //global $fileManager;
        $inputField = $this->inputField;
        $docPreview = $this->docPreview;
        $dirvar     = $this->dirvar;
        $access     = $this->access;
        $basedir    = $this->basedir;

        //$defaultHeight=70;
        $defaultWidth=120;

        $defaultHeight= ($this->iconHeight)?$this->iconHeight:70;
        $defaultWidth = ($this->iconWidth)? $this->iconWidth:120;
        
        if ($modo == 'lista') {
            $defaultHeight=20;
            $defaultWidth=20;
        }

        foreach($imagenes as $clave => $valor) {
            $Nfoto  = $clave;
            $Path	= $path;
            $Imagen = $Path .  $valor;

            $stat =  @stat( $Imagen);
            $ultimaMod = $stat['mtime'];

            $ImagenReal2= $Path. rawurlencode($valor);
            $ImagenReal= $Imagen;
            
            $hash = @md5_file($ImagenReal);
            if ($_GET['del'] != '' && $_GET['del'] == $hash) {
                unlink($ImagenReal);
                continue;
            }

    	    $pathinfo = pathinfo_utf($ImagenReal);


            // creates File Object

            $archivo   = new Archivo($pathinfo['basename'], $Path, '');
            $filesize  = $archivo->filesize;
            $exif 	   = $archivo->exif;

            // MIMETYPEs
            $mimetype= $archivo->mimetype;

                    /* SVG exception */
            if ($mimetype == 'text/xml' && strpos(strtolower($ImagenReal), ".svg"))
                $mimetype = 'image/svg';
            if (strpos(strtolower($ImagenReal), ".xls"))
                $mimetype = 'application/xls';

            $back = '';

            $display = ($modo == 'lista')?'inline':'block';
            $openLink  = '';
            $closeLink ='';
            if ($archivo->preview) {
                $uid= uniqid('img');
                $openLink  = '<a  class="boton" onMouseOver="showImage(this, \''.$ImagenReal.'\', event, \'Supracontenido\', \''.$archivo->nombre.'\',  \''.$uid.'\');" onMouseOut="cerrarVent(\''.$uid.'\');" style="padding:3px;display:'.$display.';" title="'.$valor.'" rel="lightbox[fileManager]" href="thumb.php?page=0&url='.urlencode($ImagenReal).'&ancho=650&alto=500&fs='.$filesize.'" >';
                $closeLink ='</a>';
            }

            unset ($buttons);
            if ($archivo->tipo == 'doc' ||
                $archivo->tipo == 'pdf' ||
                $archivo->tipo == 'ods' ||
                $archivo->tipo == 'odp' ||
                $archivo->tipo == 'odt' ||
                $archivo->tipo == 'xls' ||
                $archivo->tipo == 'txt' ||
                $archivo->tipo == 'dwg'
            ) {

                $buttons['view'] =   $archivo->viewerButton()->show();
            }

            if ($docPreview)  $doc='&docPreview=true';
                         else $doc='&docPreview=false';

            $miImagen = $openLink.$archivo->thumb($defaultWidth, $defaultHeight,null, '&fs='.$filesize.$doc).$closeLink;

            $resolucion = '';
            $resol = '';
            if ($archivo->width != '') {
                $resol = $archivo->width.' x '.$archivo->height;
                $resolucion = 'Resoluci&oacute;n: '.$resol;
            }

            $strFilesize= $archivo->byteConvert($filesize);

            // DOWNLOAD BUTTON
            if (strpos($access, 'd') !== false){
                    $buttons['download'] = $archivo->downloadButton();
            }

            // HISTORY BUTTON
            if (strpos($access, 'w') == true && is_dir($Path .  '.' .$valor)) {
                $history = new Html_button('', '../img/history.png');
                $history->addParameter('title', 'Historico');
                $history->addEvent('onclick', 'window.location=\'?inputField='.$inputField.'&basedir='.$basedir.'&dirvar='.$dirvar.'&dir2='.urlencode($slashdir.'.' .$valor.'/').'&DAT='.$_SESSION['DAT'].'\'');
                $buttons['hist'] = $history->show();
            }


            // Select Button
            if ($inputField != '') {

		// fix utf8 characters in files

		$pathinfo = pathinfo_utf($ImagenReal);
		$filename = $pathinfo['basename'];
                //$filename = basename($ImagenReal);

                $sel = new Html_button('', '../img/ok.png');

                $sel->addParameter('title', 'Seleccionar');
                $sel->addEvent('onclick', "fmReturnValue('$slashdir$filename')");
                $buttons['select'] = $sel->show();
            }


            // EDITOR BUTTON
            if ($archivo->tipo == 'xml' || $mimetype == 'text/xml') {
                $filename 	= $pathinfo['basename'];
                $editor 	= new Html_button('', '../img/edit.png');
                $editor->addParameter('title', 'Editar');
                $exec 	= "Histrix.loadExternalXML('Edit_$filename', 'codeEditor.php?file=$filename&dir=$basedir$slashdir');";
                $editor->addEvent('onclick', $exec);
                $buttons['edit'] = $editor->show();
            }

            if (strpos($access, 'w') !== false)
                $allowDelete= true;

            // DELETE BUTTON
            if ($allowDelete == true) {
                $filename = $pathinfo['basename'];
                $delete = new Html_button('', '../img/cancel.png');
                $delete->addParameter('title', 'Borrar');
                $delete->addEvent('onclick', "fmDelete('$slashdir', '$hash', '$basedir', '$dirvar' , '".$_SESSION['DAT']."');");
                $buttons['delete'] = $delete->show();
            }
            if ($archivo->tipo == 'doc'){
                $file= urlencode($ImagenReal);
                $exec 	= "Histrix.loadInnerXML('HTML$filename', 'docReader.php?file=$file', null, 'Documento');";

                $valor = '<a onclick="'.$exec.'"  "docReader.php?file='.$file.'" target="reader">'.$valor.'</a>';
            }

            if ($modo =='lista' ) {
                $lista  = '<tr>';
                $lista .= '<td>'.$miImagen.$valor.'</td>';
                $lista .= '<td>'.$strFilesize;
                if ($resol != '')
                    $lista .= ' ('.$resol.')';
                $lista .= '</td>';

                $lista .= '<td>';

                if ($buttons != '') {
                     foreach ($buttons as $nbut => $button){
                         $lista .= $button;
                     }
                 }

                $lista .= '</td>';

                $lista .= '<td>'.$mimetype.'</td>';
                $lista .= '<td>'.date('d/m/Y - H:i:s',$ultimaMod).'</td>';
                $lista .= '</tr>';

                $salida .=  $lista;
            }
            else {
                $foto = '<span class="polaroid2"  >';
                $foto .= '<div  style="cursor:pointer;">';
                $foto .= '<a  target="_blank" href="'.$ImagenReal2.'&fs='.$filesize.'">';
                $foto .= $miImagen;
                $foto .= '<div class="exif" style="'.$back.'" >';
                $foto .= '<a  target="_blank" href="'.$ImagenReal.'">'.utf8_decode($valor).'</a><br>';
                $foto .= $resolucion .'<br>';
                $foto .= 'Tama&ntilde;o: '.$strFilesize .'<br>';

                if ($exif)
                    $foto .= "Fecha: ".$exif['IFD0']['DateTime'].'<br>';

                if ($buttons != '')
                    foreach ($buttons as $nbut => $button){
                        $foto .= $button;
                    }

                $foto .= '</div>';
                $foto .= '</div>';
                $foto .= '</span>';
                $foto .= "\n";

                $salida .=  $foto;
            }


        };// fin del while.

        return $salida;
    }

    function  navPanel( $basepath, $slashdir){
        if ($this->modo == 'lista') {
                $openTable = '<table class="sortable resizable"  border="0" width="99%">';
                $openTable .= '<thead>';
                $openTable .= '<tr><th>Nombre</th><th>Tama&ntilde;o</th><th>Acci&oacute;n</th>';
                $openTable .= '<th>Tipo</th><th>Fecha de modificaci&oacute;n</th></tr>';
                $openTable .= '</thead>';
                $openTable .= '<tbody style="overflow:hidden;">';

                $closeTable .= '</tbody>';
                $closeTable .= '</table>';
        }
        $salida .= $openTable;
        $salida .= $this->getDirContents( $basepath, $slashdir, 0);
        $salida .= $closeTable;
        return $salida;
    }


    function miniThumb($file, $documentPreview=true) {
        $width = 40;
        $height= 35;
        $filesize= filesize($file);
        if ($documentPreview) $docPreview = '&docpreview=true';
        else $docPreview = '&docpreview=false';
        $urlfile = urlencode($file);
        $Imagen = 'thumb.php?url=' . $urlfile .'&ancho='.$width.'&alto='.$height.'&fs='.$filesize.$docPreview;
        $img = '<img class="mini" src="'.$Imagen.'"  >';
        return $img;

    }

}
?>
