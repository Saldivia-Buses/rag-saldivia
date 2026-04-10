<?php

//require ('../lib/fpdf/fpdf.php');
if (!defined('FPDF_FONTPATH'))
    define('FPDF_FONTPATH', '../../lib/fpdf/font/');

class Histrix_pdf extends FPDF {

    var $B;
    var $I;
    var $U;
    var $HREF;
    var $titulo;
    var $anchoPagina;
    var $offsetX;
    var $offsetY;
    var $maxY;
    var $sincab;
    var $fontsize;

    function __construct($orientation = 'P', $unit = 'mm', $format = 'A4') {
        $this->defaultFont = 'helvetica';
        $this->fontsize = 8;
        $this->pageNumber = 0;
        $this->headerMargin = 27;
        $db = $_SESSION["db"];
        $this->user = $_SESSION["usuario"];

        // Get utf8 internationalization strings
        $registry =& Registry::getInstance();
        $i18n = $registry->get('i18n');
        $this->i18n  = array_map("utf8_encode", $i18n);


        $datosbase = Cache::getCache('datosbase' . $db);
        if ($datosbase === false) {
            $config = new config('config.xml', '../database/', $db);
            $datosbase = $config->bases[$db];
            Cache::setCache('datosbase' . $db, $datosbase);
        }
        $this->datosbase = $datosbase;

        //Call parent constructor
        parent::__construct($orientation, $unit, $format, true, 'UTF-8');
        //Initializationheader
        $this->B = 0;
        $this->I = 0;
        $this->U = 0;
        $this->HREF = '';
        $this->bottomMargin = -12;

        if ($orientation == 'P') {
            $this->offsetX = 4;
            $this->offsetY = 5;
            $this->margenX = 3;
        } else {
            $this->offsetX = 3;
            $this->offsetY = 8;
            $this->margenX = 4;
        }
        $this->anchoPagina = $this->w - ($this->offsetX / 2) - ($this->margenX * 2); // margenes simetricos
    }

    // PRINT LABEL
    // TODO: redo Method
    function labelXY($objCampo, $nolabel, $MisDatos, $customValue = '') {
        $defaultFont = $this->defaultFont;
        $font = $defaultFont;
        $this->SetFillColor(255, 255, 255);
        $this->SetDrawColor(192, 192, 192);
        $this->SetTextColor(0);
        $this->SetFont('');
        $fill = 0;
        $transform = false;
        $estiloFont = '';
        $sinborde = false;
        $etiqueta = utf8_decode($objCampo->Etiqueta);

        if ((isset($objCampo->PDFnolabel) && $objCampo->PDFnolabel == 'true') || $etiqueta == '') {
            $nolabel = 'true';
        }

        // almaceno posicion actual
        $xold = $this->getX();
        $yold = $this->getY();
        //$valor = utf8_decode($objCampo->valor);

        $valor = $objCampo->valor;

        if ($customValue != '')
            $valor = $customValue;

        $sizeOrig = $this->fontsize;
        $size = $sizeOrig;


        if ($objCampo->TipoDato == 'date' && $valor != '')
            $valor = date("d/m/Y", strtotime($valor));

        $FieldOffsetY = (isset($objCampo->offsetY)) ? $objCampo->offsetY : 0;
        $FieldOffsetX = (isset($objCampo->offsetX)) ? $objCampo->offsetX : 0;
        $posY = $objCampo->posY + $FieldOffsetY + $this->offsetY;
        $posX = $objCampo->posX + $FieldOffsetX + $this->offsetX;

        if ($objCampo->posY == -1)
            $posY = $this->maxY;
        if ($objCampo->posY == 0)
            $posY = $this->lastY;

        if (substr($objCampo->posY, 0, 1) == '+')
            $posY = $this->maxY + $objCampo->posY;

        $this->lastY = $posY;
                  
	
        if (isset($objCampo->contExterno) && $objCampo->esTabla) {

            if (isset($objCampo->PDFsize)) {
                $this->SetFontSize($objCampo->PDFsize);
                $this->fontsize = $objCampo->PDFsize;
            }

            $UI = 'UI_' . str_replace('-', '', $objCampo->contExterno->tipo);
            $abmDatosDet = new $UI($objCampo->contExterno);
            $abmDatosDet->showTablaInt('noecho', '', '', 'nocant');

            if ($abmDatosDet->registros > 0) {

                $this->SetXY($posX, $posY);
                if ($etiqueta != '' && $nolabel != 'true') {
                    $this->Cell(0, 0, $etiqueta);
                    $this->ln(1);
                    $this->maxY = max($this->maxY, $this->GetY() /* $posY */
                    );
                    $posY = $this->maxY;
                }
                $opImpresion = '';
                if (isset($objCampo->Parametro['pdflineas']) && $objCampo->Parametro['pdflineas'] == 'false')
                    $opImpresion['lineas'] = false;
                if (isset($objCampo->Parametro['pdftitulo']) && $objCampo->Parametro['pdftitulo'] == 'false')
                    $opImpresion['titulo'] = false;
                $fontsize = $this->fontsize;
                if (isset($objCampo->titulo)) {
                    $this->Multicell(0, ($this->fontsize / 2), utf8_decode($objCampo->titulo));
                    $this->ln(1.5);
                }

                $pdfwidth = (isset($objCampo->Parametro['pdfwidth'])) ? $objCampo->Parametro['pdfwidth'] : null;
                $pdfwidth = (isset($objCampo->PDFwidth)) ? $objCampo->PDFwidth : $pdfwidth;

		// Use Custom Print Method
		if (method_exists($abmDatosDet, 'pdf')){
		    // use Custom pdfPrint Method
		    $abmDatosDet->pdf($this, $fontsize, $opImpresion, $pdfwidth,  $objCampo->posx);

		}
		else{
                  if ($objCampo->contExterno->tipoAbm == 'chart') {

                    $this->showGraficos($objCampo->contExterno);
                  } else {
		    // fallback Print
                    $this->impTabla($objCampo->contExterno, null, $opImpresion, $pdfwidth, $objCampo->PDFsize, $objCampo->posx);
                  }
                }
                $this->maxY = max($this->maxY, $this->GetY());
                $posY = $this->maxY;

                $this->fontsize = $fontsize;
            }
        } else {

            //	$valor = $objCampo->valor;
            $valor = utf8_decode($objCampo->valor);

            if ($customValue != '')
                $valor = $customValue;

            switch ($objCampo->TipoDato) {
                case "integer" :
                    break;
                case "decimal" :
                case "numeric" :
                case "custom_numeric" :
                    $precision = ($objCampo->numberPrecision != '')? $objCampo->numberPrecision:2;
                    if (is_numeric($valor))
                        $valor = number_format($valor, $precision, ',', '.');
                    break;
                case "date" :
                    if ($valor == '0000-00-00')
                        $valorMed = '';
                    else
                        $valorMed = 'XX/XX/XXXX';
                    break;
            }

            if ($objCampo->TipoDato == 'date' && strpos($valor, '-') == 4)
                $valor = date("d/m/Y", strtotime($valor));

            /* FORMATEO EXPLICITO */
            $formatoCampo = (isset($objCampo->Formato)) ? $objCampo->Formato : '';
            if ($formatoCampo != '') {
                // las fechas se formatean diferente
                if ($objCampo->TipoDato == 'date' || $objCampo->TipoDato == 'time') {
                    $valor = date($formatoCampo, strtotime($objCampo->valor));
                    $size = strlen($valor);
                } else {
                    $valor = sprintf($formatoCampo, $valor);
                    $size = strlen(sprintf($formatoCampo, $valor));
                }
            }

            if (isset($objCampo->opcion)) {
                $valor = $objCampo->opcion[$valor];
            }
            if (is_array($valor)) {
                $valor = utf8_decode(current($valor));
            }
            //$valor = utf8_decode($valor);

            if (isset($objCampo->aletras) && $objCampo->aletras == true) {
                $valor = NumeroALetras($valor);
            }
            if (isset($objCampo->prefijo))
                $valor = $objCampo->prefijo . $valor;

            // Label of label :)
            if ($etiqueta != '')
                $datos['datos'][0][] = $etiqueta;

            // Label Data
            $datos['datos'][0][] = $valor;

            //$datos['atributos'][0][1][] = 'sinborde';
            //$datos['atributos'][0][1][] = 'left';


            if (isset($objCampo->PDFimg) && $objCampo->PDFimg == 'true') {
                //$datos['atributos'][0][1][] = 'img='.$valor;
                $datos['atributos'][0][1][] = 'img'; // Datos
            }

            if (isset($objCampo->PDFscaleX)) {
                //Start Transformation
                $this->StartTransform();
                $this->ScaleX($objCampo->PDFscaleX);
                $transform = true;
            }

            // CUSTOM ROTATION SUPPORT
            if (isset($objCampo->PDFrotate)) {
                //Start Transformation
                $this->StartTransform();
                $this->Rotate($objCampo->PDFrotate);
                $transform = true;
            }


            if (isset($objCampo->PDFsize)) {
                $this->SetFontSize($objCampo->PDFsize);
                $size = $objCampo->PDFsize;
                $h = ($this->fontsize / 2);
                $sizeesp = true;
            }

            if (isset($objCampo->PDFfont)) {

                if ($this->fontsAdded[$objCampo->PDFfont] != true)
                    $this->AddFont($objCampo->PDFfont, '', $objCampo->PDFfont . '.php');
                $this->fontsAdded[$objCampo->PDFfont] = true;
                $this->SetFont($objCampo->PDFfont);
            }

            if (isset($objCampo->PDFcolor)) {
                $textrgb = hex_to_rgb($objCampo->PDFcolor);
                $this->SetTextColor($textrgb['red'], $textrgb['green'], $textrgb['blue']);
            }

            if (isset($objCampo->PDFbackgroundColor) && $objCampo->PDFbackgroundColor != '') {
                $textrgb2 = hex_to_rgb($objCampo->PDFbackgroundColor);
                $this->SetFillColor($textrgb2['red'], $textrgb2['green'], $textrgb2['blue']);
                $fill = true;
            }

            //$datos['atributos'][0][0][]='sinborde';

            $w = $this->setAnchoCol($datos);
            if (isset($objCampo->lblsize))
                $w[0] = $objCampo->lblsize;
            if (isset($objCampo->pdfancho))
                $w[1] = $objCampo->pdfancho;

            //	$this->WriteTable($datos, $w, $posX, $posY, $objCampo);
            //   return;
            // Print using multicell
//            $valor = $this->ReplaceHTML($valor);

            if (is_utf8($valor))
                $valor = utf8_decode($valor);

            $nb = 0;

            $row[] = $etiqueta;
            $row[] = $valor;


            for ($i = 0; $i < 2; $i++)
                $nb = max($nb, $this->NbLines($w[$i], trim($row[$i])));

            $h = ($this->fontsize / 2) * $nb;

            $align = 'L';
            //$borde = 'B';
            $estilo = 'F';
            $soloborde = true;
            $this->SetDrawColor(192, 192, 192);

            $align = (isset($objCampo->PDFalign)) ? $objCampo->PDFalign : 'L';

	
            if (isset($objCampo->PDFnoborder)) {
                $sinborde = true;
                $borde = '';
            }
            $borde = (isset($objCampo->PDFborder)) ? $objCampo->PDFborder : $borde;
            if (isset($objCampo->pdffill)) {
                $this->SetFillColor($objCampo->pdffill);
            }

            $lineH = ( $this->fontsize / 2 ) * $nb;

            // Label

            if ($nolabel != 'true') {
                $this->SetXY($posX, $posY);
                $pdffill = (isset($objCampo->pdffill) ) ? $objCampo->pdffill : '';
                if ($pdffill == '')
                    $this->SetFillColor(240, 240, 240);



                $nb1 = $this->NbLines($w[0], $etiqueta);
                $h1 = ($lineH / $nb1);
		
		
		$checkPage = (isset($objCampo->checkPageBreak) && $objCampo->checkPageBreak == 'false')?0:$this->CheckPageBreak($h1);
                if ($checkPage) {
                    $posY = $this->headerMargin;
                    $this->SetY($posY);
                }
                if (!($sinborde)) {
                    if ($soloborde)
                        $this->Rect($posX, $posY, $w[0], $h1, $estilo);
                    else
                        $this->Rect($posX, $posY, $w[0], $h1, $estilo);
                } 

                if (isset($objCampo->PDFitalic) && ( $objCampo->PDFitalic == 'label' ||
                        $objCampo->PDFitalic == 'both')) {

                    $estiloLabelFont .= 'I';
                }
                if (isset($objCampo->PDFbold) && ( $objCampo->PDFbold == 'label' ||
                        $objCampo->PDFbold == 'both')) {
                    $estiloLabelFont .= 'B';
                }
                if ($estiloLabelFont != '')
                    $this->SetFont(null, $estiloLabelFont, $size);

                $this->MultiCell($w[0], $h1, $etiqueta, $borde, $align, $fill);
                $this->maxY = max($this->maxY, $this->GetY());
                $this->SetXY($posX + $w[0], $posY);
            }
            else {
                if (isset($objCampo->lblsize))
                    $w[1] = $objCampo->lblsize;
                $w[0] = 0;
//                $fill = 0;
                $this->SetXY($posX, $posY);
            }
            // Data printing
            switch ($objCampo->TipoDato) {
                case "integer" :
                case "decimal" :
                    $align = 'R';
                    break;
                case "numeric" :
                    $align = 'R';
                    //                    $valor = number_format($valor, 2, '.', ',');
                    break;
                case "date" :
                case "time" :
                    $align = 'L';

                    break;
                case "editor":
                case "simpleditor":
                    $this->SetXY($posX + $w[0], $posY);
                    $valor = html_entity_decode($valor, ENT_QUOTES);
                    // FCKEditor page Breack
                    $valor = str_replace('<div style="page-break-after: always;">', '<page></page><div>', $valor);
                    $valor = str_replace('&hellip;', '.', $valor);
                    $this->WriteHTML2($valor, $posX,'');
                    return;
                    break;
                case "dir":
                    /* browse dir */
                    //include ('./FileManager_class.php');
                    $basePath = '../database/' . $_SESSION['datapath'] . 'xml/';
                    $basedir = $valor;
                    $fileManager = new FileManager($slashdir, $basePath, $access);
                    $fileManager->basedir = $basedir;
                    $images = $fileManager->getDirContents($basePath, $basedir, 1);

                    $margin = 2;
                    $cant = count($images);

                    $firstPosX = $posX;
                    
                    if (is_array($images))
                        foreach ($images as $n => $imageurl) {
                            $url = $basePath . $basedir . $imageurl;
//                            echo $basePath.'_'. $basedir.'_'.$imageurl;
                            $file = new Archivo($imageurl, '', '', $objCampo);
                            $x_resolution = 1000;
                            $dpi = 100;
                            if ($objCampo->watermark != '') {
                                $file->watermark = $MisDatos->getCampo($objCampo->watermark)->valor;
                            }
                            //$file->watermark = 'MARCA DE AGUA';
                                              /*
			    if (isset($objCampo->imageCrop)){
				$crop = implode(',',$objCampo->imageCrop);
				$imageCrop['w'] = $crop[1];
				$imageCrop['h'] = $crop[2];
				$imageCrop['x'] = $crop[3];
				$imageCrop['y'] = $crop[4];
			    }
			    
                                                */
                            $image      = $file->toImage($url, $x_resolution, 0, 0, $dpi);
                            $Imgsize    = ($objCampo->imageWidth != '') ? $objCampo->imageWidth : 50;
			    
			    if (is_file($image))
                            $sizesarray = $this->Image($image, $posX + $w[0] + $this->offsetX + $imageWidthMargin, $posY + $imageHeightMargin, $Imgsize, null, 'jpg', '', false);

                            if ($n <= $cant) {
				// set Current Position
                                $this->setY($posY);
                                $this->setX($posX);


				// get max height of image on current Row
                                $rowmaxheight = max($rowmaxheight, $sizesarray['HEIGHT'] );
				    // if image fits horizontally then its ok
                                if (($posX + $sizesarray['WIDTH'] + $margin) < $this->anchoPagina){
    




                                }
                                else {
                            	    // image dont fit horizontally then move to new line
                                    if ($this->CheckPageBreak($sizesarray['HEIGHT'] + $posY )) {
					// new page
                                        $posY = $this->headerMargin;	// move to header
                                        $posX = $firstPosX ; 		// reset horizontal column
                                    }
                                    else {
					// new row
	                                $posY += $rowmaxheight + $margin;
                                        $posX  = $firstPosX ;
                                    }
				    // reset row height
                                    $rowmaxheight = 0;

                                }

				// draw image
			    if (is_file($image))

                                $this->Image($image, $posX + $w[0] + $this->offsetX + $imageWidthMargin, $posY + $imageHeightMargin, $Imgsize, null, 'jpg', '', true);
				// move to next position horizontally
                                $posX += $sizesarray['WIDTH'] + $margin;
                                
                            }
                        }
                    return;
                    break;
                case "grafico" :
//                                $this->setY($posY);

                    $align = 'C';
                    /*
                      $max = $MisDatos->TablaTemporal->getMax($nombreCampo);
                      $min = $MisDatos->TablaTemporal->getMin($nombreCampo);
                     */
                    if (isset($objCampo->max)) {
                        $max = $ObjCampo->max;
                        $min = 0;
                        $showval = 1;
                    } else {
                        $max = $MisDatos->TablaTemporal->getMax($objCampo->NombreCampo);
                        $min = $MisDatos->TablaTemporal->getMin($objCampo->NombreCampo);
                    }
                    $Imgsize = ($objCampo->imageWidth != '') ? $objCampo->imageWidth : 50;

                    $clase = 'grHoriz';
                    $anchoIMG = $Imgsize;
                    $altoIMG = ($objCampo->imageHeight != '') ? $objCampo->imageHeight : 20;
                    $scale = ($objCampo->imageScale != '') ? $objCampo->imageScale : 1.5;

                    $uid = uniqid();
                    $imagen = '../database/' . $_SESSION['datapath'] . 'tmp/' . $uid;

                    if ($objCampo->grafico == 'barcode' || $objCampo->encode != '' )
                        $clase = 'barcode';

                    $img = new Graficar($clase, $anchoIMG, $altoIMG, $imagen, 'jpg');

                    switch ($clase) {
                        case "grHoriz":
                            $img->crearImagen();
                            $img->grHoriz($valor, $min, $max, $c1, $c2, $showval);
                            break;
                        case "barcode":
                            if ($valor != '') {
                                $encode = ($objCampo->encode != '') ? $objCampo->encode : 'I25';
                                $img->barcode($valor, $altoIMG, $scale, $imagen, $encode);
                            }
                            break;
                    }
                    if ($img->imagen != '')
                        @imagejpeg($img->imagen, $imagen . '.jpg');

                    if ($valor != ''){
		        if (is_file($imagen . '.jpg'))
                	    $this->Image($imagen . '.jpg', $posX + $w[0] + $this->offsetX + $imageWidthMargin, $posY + $imageHeightMargin, $Imgsize);
		    }
                    if ($img->imagen)
                        imagedestroy($img->imagen . '.jpg');

                    if ($valor != '')
                        @unlink($imagen);
                        
                    return;
                    break;



                case "file":
                    $file = new Archivo($valor, '', '', $objCampo);
                    if ($valor != '') {
                        //		        $this->SetXY($posX + $w[0], $posY);
                        $link = $objCampo->url;
                        // Path for uploading files
                        //$link .= $MisDatos->path.'/';
                        if ($MisDatos->path != '') {
                            $filePath = '../database/' . $_SESSION['datapath'] . 'xml/' . $MisDatos->path;
                            $link = $filePath . '/';
                        }

                        // Add Object Path
                        if ($objCampo->path != '') {
                            $filePath = '../database/' . $_SESSION['datapath'] . 'xml/' . $MisDatos->path;

                            $link = $filePath;
                            $Objurl = $MisDatos->getCampo($objCampo->path)->valor;
                            $link .= $Objurl . '/';
                        }
                        $link .=$file->link;

                        if (($file->imagen || $file->preview ) && $file->svg != true) {
                            $link = utf8_encode($link);


                            if (is_file($link)) {
                                $img = $link;
                                $width = $objCampo->Size;
                                $Imgsize = ($objCampo->imageWidth != '') ? $objCampo->imageWidth : $width;
                                if ($Imgsize == 0)
                                    $Imgsize = 5;
                                /* Resize Image so the PDf is lighter */

                                $source_pic = $file->toImage($link);
                                //$source_pic = $link;

                                $max_width = 1600;
                                $max_height = 1200;

                                $dataPath = $_SESSION['datapath'];
                                if ($dataPath != '') {
                                    $tmpbase = '../database/' . $dataPath;
                                }
                                $destination_pic = $tmpbase . '/tmp/' . uniqid('pdfimg') . '.jpg';

                                $src = imagecreatefromjpeg($source_pic);

                                list($width, $height) = getimagesize($source_pic);
                                // Just DownSize Images
                                if ($width < $max_width)
                                    $max_width = $width;
                                if ($height < $max_height)
                                    $max_height = $height;

                                $x_ratio = $max_width / $width;
                                $y_ratio = $max_height / $height;
                                if (($width <= $max_width) && ($height <= $max_height)) {
                                    $tn_width = $width;
                                    $tn_height = $height;
                                } elseif (($x_ratio * $height) < $max_height) {
                                    $tn_height = ceil($x_ratio * $height);
                                    $tn_width = $max_width;
                                } else {
                                    $tn_width = ceil($y_ratio * $width);
                                    $tn_height = $max_height;
                                }
                                $tmp = imagecreatetruecolor($tn_width, $tn_height);
                                imagecopyresampled($tmp, $src, 0, 0, 0, 0, $tn_width, $tn_height, $width, $height);

                                imagejpeg($tmp, $destination_pic);
                                imagedestroy($src);
                                imagedestroy($tmp);
                                $imageHeightMargin = 1;
				if (is_file($destination_pic)){

                                    $sizesarray = $this->Image($destination_pic, $posX + $w[0] + $this->offsetX + $imageWidthMargin, $posY + $imageHeightMargin, $Imgsize, null, 'jpg');
				}
                                unlink($destination_pic);

                                // $sizesarray['WIDTH'];
                                $h = $sizesarray['HEIGHT'] + $imageHeightMargin * 2;

                                $nb = ($this->fontsize / 2) / $h;
                            }
                            return;
                        }
                    }
                    break;
            }

            if (isset($objCampo->opcion))
                $align = 'L';



            $borde = '';
            if (isset($objCampo->PDFborder))
                $borde = $objCampo->PDFborder;

            if (isset($objCampo->PDFitalic) && ( $objCampo->PDFitalic == 'true' ||
                    $objCampo->PDFitalic == 'both')) {
                $estiloFont .= 'I';
            }

            if (isset($objCampo->PDFbold) && ( $objCampo->PDFbold == 'true' ||
                    $objCampo->PDFbold == 'both')) {
                $estiloFont .= 'B';
            }

            $this->SetXY($posX + $w[0], $posY);

            $printValue = true;
            if (isset($objCampo->color) && $objCampo->color == 'true') {
                $rgb = hex_to_rgb($valor);
                if ($rgb) {
                    $this->SetFillColor($rgb['red'], $rgb['green'], $rgb['blue']);
                    $this->SetDrawColor($rgb['red'], $rgb['green'], $rgb['blue']);
                }
                $this->Rect($posX + $w[0], $posY, 5, 5, 'DF');
                $printValue = false;
            }

            if ($objCampo->TipoDato == 'check') {
                $this->SetDrawColor(0, 0, 0);
                $this->CheckBox(($posX + $w[0] + 1), ($posY + 1), $valor);
                $this->SetDrawColor(192, 192, 192);
                $printValue = false;
            }
//            $this->SetFont($defaultFont, $estiloFont, $size);
            $this->SetFont(null, $estiloFont, $size);


            if ($printValue) {
                $h = $lineH;
                if ($nb > 1)
                    $h = $lineH / $nb;

				$checkPage = (isset($objCampo->checkPageBreak) && $objCampo->checkPageBreak == 'false')?0:$this->CheckPageBreak($h);
                if ($checkPage) {
                    $posY = $this->headerMargin;
                    $this->SetY($posY);
                }
                //$align = 'R';

                $this->MultiCell($w[1], $h, $valor, $borde, $align, $fill);
            }

            if ($transform)
                $this->StopTransform();

	    if ($objCampo->PDFrestorePosition == "true"){
		$this->SetXY($xold, $yold);
	    }

            $this->SetFont($defaultFont, $estiloFont, $size);
            $this->SetFont('');

            //Put the position to the right of the cell
            $size = $sizeOrig;
            $this->SetFontSize($size);
        }

        if (isset($objCampo->PDFbox) && $objCampo->PDFbox == "true") {
            $this->Rect($posX, $posY, $w[0] + $w[1], $this->maxY - $posY);
        }
        // restauro posicion
        $this->maxY = max($this->maxY, $this->GetY());
        if (isset($lineH))
            $this->Ln($lineH);
            
	return $this->GetY();            
    }

    function doublemax($mylist) {
        $maxvalue = max($mylist);
        while (list($key, $value) = each($mylist)) {
            if ($value == $maxvalue)
                return array("key" => $key, "value" => $value);
        }
    }

    // ajusto los anchos de las columnas
    function ajustoAnchos($anchoColumna, $anchoTabla=null) {
        if (!is_array($anchoColumna))
            return 0;
        $anchofilafin = array_sum($anchoColumna);
        $anchopag = $this->anchoPagina - $this->offsetX / 2;
        if ($anchoTabla != null) {
            $anchopag = $anchoTabla;
        }

        // Achico la Tabla
        if ($anchofilafin > $anchopag) {
            while ($anchofilafin >= $anchopag) {

                $colarray = $this->doublemax($anchoColumna);
                $col = $colarray['key'];
                $anchoColumna[$col] -= 1;
                $anchofilafin = array_sum($anchoColumna);
            }
        } else {
            // calculo proporciones
            $noresizecant = (isset($this->noResize)) ? count($this->noResize) : 0;

            $cant = count($anchoColumna) - $noresizecant;

            if ($cant == 0)
                $cant = 1;
            $difs = ($anchopag - $anchofilafin) / $cant;

            foreach ($anchoColumna as $nom => $ancho) {
                //$porcelda = ($ancho / $anchofilafin) * 100;
                //$anchoColumna[$nom] = ($anchopag / 100) * $porcelda;
                if (!isset($this->noResize[$nom]))
                    $anchoColumna[$nom] = $ancho + $difs;
            }
        }
        return $anchoColumna;
    }

	// check if a row is empty
	function isRowEmpty($row){
		$acuval = '';
		$return = true;
        if ($row) {
            foreach ($row as $name => $value) {
                $acuval .= trim($value);
            }
            if ($acuval != '') $return = false;
        }		
        return $return;
        
	}

    // Imprime la cabecera de las tablas
    function impCabTabla($row, $anchoColumna, $arraycampos=false, $fontsize=10, $checkPage = true) {

        // Return if no header is present
		if ($this->isRowEmpty($row)){
			return false;
		}
			

        $inX = $this->getX();
        $nb = 0;
        $midFontSize = $fontsize / 2;
        if ($row) {
            $this->SetFont($this->defaultFont, 'B', $fontsize);
            foreach ($row as $nom => $valor) {
                $nb = max($nb, $this->NbLines($anchoColumna[$nom], trim($valor)));
            }
        }

        $h = ($fontsize / 2) * $nb;

        $align = 'C';
        $estilo = 'DF';
        //$this->SetDrawColor(0, 0, 0);
        $this->SetDrawColor(192, 192, 192);
        $this->SetFillColor(230, 230, 230);
        $print = true;
        if ($row) {
            if ($checkPage) {
                $salto = $this->checkPageBreak($h);
            }
            // $this->setX($posX);

            if (($salto)) {
                //$this->Ln(5);
                //   $this->setX($posX);
                if ($printCab == true)
                    $this->impCabTabla($cabecera, $anchoColumna, null, $fontsize);
            }

            foreach ($row as $nom => $valor) {
                $valor = utf8_decode($valor);
                if ($arraycampos != '') {

                    if ($valor != '') {
                        $this->SetFillColor(230, 255, 214);
                        $print = true;
                    }
                    else
                        $print = false;

                    $ObjCampo = $arraycampos[$nom];
//                    if ($ObjCampo->PDFlabel == 'false') continue;

                    switch ($ObjCampo->TipoDato) {
                        case "integer" :
                            $align = 'R';
                            break;
                        case "decimal" :
                        case "numeric" :
                        case "custom_numeric" :
                            $align = 'R';
                            $precision = ($ObjCampo->numberPrecision != '')? $ObjCampo->numberPrecision:2;
                            if (is_numeric($valor))
                                $valor = number_format($valor, $precision, ',', '.');
                            break;


                        case "date" :
                        case "time" :
                            $align = 'C';
                            break;
                    }
                }
                $x = $this->GetX();
                $y = $this->GetY();

                if ($print) {
                    $this->Rect($x, $y, $anchoColumna[$nom], $h, $estilo);
                    $this->MultiCell($anchoColumna[$nom], ($midFontSize), trim($valor), 0, $align);
                }
                //Put the position to the right of the cell
                $this->SetXY($x + $anchoColumna[$nom], $y);
            }
        }
        $this->Ln($h);
        $this->SetFont($this->defaultFont,'');

        $this->setX($inX);
    }

    function getAnchoTabla($MisDatos, $Tablatemp) {

        if ($Tablatemp != '')
            foreach ($Tablatemp as $orden => $row) {
                foreach ($row as $nombreCampo => $valor) {

                    if (!isset($tmpObj[$nombreCampo])) {
                        $tmpObj[$nombreCampo] = $MisDatos->getCampo($nombreCampo);
                    }
                    $ObjCampo = $tmpObj[$nombreCampo];
                    if (!is_object($ObjCampo))
                        continue;
                    if (isset($ObjCampo->Oculto) && ($ObjCampo->Oculto || $ObjCampo->Oculto == 'true'))
                        continue;
                    if (isset($ObjCampo->print) && $ObjCampo->print == 'false')
                        continue;
                    if (isset($ObjCampo->noshow) && $ObjCampo->noshow == 'true')
                        continue;

                    if (isset($ObjCampo->noEmpty) && $ObjCampo->noEmpty == 'true' && !isset($MisDatos->hasValue[$ObjCampo->NombreCampo])) {
                        continue;
                    }

                    if (isset($ObjCampo->PDFcolWidth)) {
                        $anchoColumna[$nombreCampo] = $ObjCampo->PDFcolWidth;
                        $ObjCampo->resize = 'false';
                    } else {



                        // Si tiene opciones de un combo
                        if (isset($ObjCampo->opcion)) {
                            $valop = (isset($ObjCampo->valop)) ? $ObjCampo->valop : '';

                            if (count($ObjCampo->opcion) > 0 && $ObjCampo->TipoDato != "check" && $valop != 'true') {
                                if (isset($ObjCampo->opcion[$valor]))
                                    $valor = $ObjCampo->opcion[$valor];
                                if (is_array($valor))
                                    $valor = current($valor);
                            }
                        }
                        $valorMed = $valor;

                        switch ($ObjCampo->TipoDato) {
                            case "integer" :
                                break;
                            case "decimal" :
                            case "numeric" :
                            case "custom_numeric" :
                                $precision = ($ObjCampo->numberPrecision != '')? $ObjCampo->numberPrecision:2;
                                if (is_numeric($valor))
                                    $valorMed = number_format($valor, $precision, ',', '.');
                                break;
                            case "date" :
                                if ($valor == '0000-00-00')
                                    $valorMed = '';
                                break;
                        }
                        /* FORMATEO EXPLICITO */
                        if (isset($ObjCampo->Formato)) {
                            $formatoCampo = $ObjCampo->Formato;

                            if ($formatoCampo != '') {
                                // las fechas se formatean diferente
                                if ($ObjCampo->TipoDato == 'date' && isset($ObjCampo->valor) || $ObjCampo->TipoDato == 'time') {
                                    $valor = date($formatoCampo, strtotime($ObjCampo->valor));
                                } else {
                                    $valor = sprintf($formatoCampo, $valor);
                                }
                                $valorMed = $valor;
                            }
                        }
                        $ancho = 0;

                        if (!isset($anchoTit[$nombreCampo])) {

                            $this->SetFont($this->defaultFont, 'B', $this->fontsize);
                            $valorEt = trim($ObjCampo->Etiqueta);
                            $valorEt = html_entity_decode(strip_tags($valorEt), ENT_QUOTES);

                            /*                             * */
                            $anchoTit[$nombreCampo] = $this->GetStringWidth(trim($valorEt)) + 3;
                            $this->SetFont($this->defaultFont);
                        }
                        $anchoTxt = $this->GetStringWidth(trim($valorMed)) + 3;
                        $ancho = max($anchoTxt, $anchoTit[$nombreCampo], $ancho);

                        /* if ($ObjCampo->size != '') {
                          $txt = str_pad('x', $ObjCampo->size , "x");
                          //$ancho = $this->GetStringWidth(trim($txt));
                          } */
                        switch ($ObjCampo->TipoDato) {
                            case "grafico" :
                                $ancho = 10;
                                break;
                        }

                        // Para el caso de tablas en tablas (picture in picture je)

                        if (isset($ObjCampo->contExterno) && ( isset($ObjCampo->esTabla) && $ObjCampo->esTabla)) {
                            $UI = 'UI_' . str_replace('-', '', $ObjCampo->contExterno->tipo);
                            if ($ObjCampo->paring != '') {
                                foreach ($ObjCampo->paring as $destinodelValor => $origendelValor) {
                                    $valorDelCampo = $row[$origendelValor['valor']];
                                    if ($valorDelCampo == '')
                                        $valorDelCampo = '0';
                                    if ($MisDatos->getCampo($origendelValor['valor'])->TipoDato == 'varchar')
                                        $comillas = '"';
                                    $operador = '';
                                    $operador = $origendelValor['operador'];
                                    if ($operador == '')
                                        $operador = '=';
                                    $reemplazo = '';
                                    $reemplazo = $origendelValor['reemplazo'];
                                    if ($reemplazo != 'false')
                                        $reemplazo = 'reemplazo';
                                    else
                                        $reemplazo = '';

                                    $ObjCampo->contExterno->addCondicion($destinodelValor, $operador, $comillas . $valorDelCampo . $comillas, 'and', $reemplazo, true);

                                    $ObjCampo->contExterno->setCampo($destinodelValor, $valorDelCampo);
                                    $ObjCampo->contExterno->setNuevoValorCampo($destinodelValor, $valorDelCampo);
                                }
                            }
                            $abmDatosDet = new $UI($ObjCampo->contExterno);

                            $abmDatosDet->showTablaInt('noecho', '', '', 'nocant');
                            $Tablatemp2 = $ObjCampo->contExterno->TablaTemporal->datos();

                            if (is_array($Tablatemp2))
                                $ancho = array_sum($this->getAnchoTabla($ObjCampo->contExterno, $Tablatemp2)) + count($Tablatemp2[0]);
                        }

                        $arrayCampos[$nombreCampo] = $ObjCampo;
                        //  if ($ObjCampo->Parametro['noshow'] == 'true'  || $ObjCampo->noshow == 'true' ) continue;
                        //  if ($ObjCampo->Parametro['print']  == 'false' || $ObjCampo->print  == 'false') continue;
                        if (isset($ObjCampo->colstyle) && strpos('_' . $ObjCampo->colstyle, 'display:none'))
                            continue;

                        if (isset($anchoColumna[$nombreCampo]))
                            $anchoColumna[$nombreCampo] = max($ancho, $anchoColumna[$nombreCampo]);
                        else
                            $anchoColumna[$nombreCampo] = $ancho;
                    }
                }

            }

        return $anchoColumna;
    }

    function getAltoFila($row, $arrayCampos, $anchoColumna, $MisDatos, $printCab=null) {
        $nb = 0;

        foreach ($row as $nombreCampo => $valor) {
            if (isset($arrayCampos[$nombreCampo])) {
                $ObjCampo = $arrayCampos[$nombreCampo];

                switch ($ObjCampo->TipoDato) {
                    case "file" :
                        //$file = new Archivo($valor, $this->Datos->path, '../archivos/');
                        $file = new Archivo($valor, '', '', $objCampo);
                        if ($valor != '') {
                            $link = $ObjCampo->url;

                            // Path for uploading files
                            //$link .= $MisDatos->path.'/';
                            if ($MisDatos->path != '') {
                                //$filePath= '../xml/'.$_SESSION['db'].'/'.$MisDatos->path;
                                $filePath = '../database/' . $_SESSION['datapath'] . 'xml/' . $MisDatos->path;
                                $link = $filePath . '/';
                            }

                            // Add Object Path
                            if ($ObjCampo->path != '') {
                                //$filePath= '../xml/'.$_SESSION['db'].'/'.$MisDatos->path;
                                $filePath = '../database/' . $_SESSION['datapath'] . 'xml/' . $MisDatos->path;
                                $link = $filePath;

                                $Objurl = $MisDatos->getCampo($ObjCampo->path)->valor;
                                $link .= $Objurl . '/';
                            }
                            $link .=$file->link;
                        }
                        break;
                    case "simpleditor" :
                    case "editor":
                        $valor = str_ireplace('<br>', "\n", $valor);
                        $valor = str_ireplace('&nbsp;', " ", $valor);

                        $valor = strip_tags($valor);
                        break;
                }

                if ($ObjCampo->TipoDato == 'grafico' && $ObjCampo->grafico == 'barcode')
                    $nb = max($nb, 1.7);

                if (isset($ObjCampo->contExterno) && isset($ObjCampo->esTabla) && $ObjCampo->esTabla) {
                    $UI = 'UI_' . str_replace('-', '', $ObjCampo->contExterno->tipo);

		            unset($abmDatosDet);
                    $abmDatosDet = new $UI($ObjCampo->contExterno);


                    // copy pasted block, replace by method
                    if ($ObjCampo->paring != '') {

                        foreach ($ObjCampo->paring as $destinodelValor => $origendelValor) {
                        
                            $valorDelCampo = $row[$origendelValor['valor']];
                            if ($valorDelCampo == '')
                                $valorDelCampo = '0';
                            if ($MisDatos->getCampo($origendelValor['valor'])->TipoDato == 'varchar')
                                $comillas = '"';
                            $operador = '';
                            $operador = $origendelValor['operador'];
                            if ($operador == '')
                                $operador = '=';
                            $reemplazo = '';
                            $reemplazo = $origendelValor['reemplazo'];
                            if ($reemplazo != 'false')
                                $reemplazo = 'reemplazo';
                            else
                                $reemplazo = '';

                            $ObjCampo->contExterno->addCondicion($destinodelValor, $operador, $comillas . $valorDelCampo . $comillas, 'and', $reemplazo, false);

                            $ObjCampo->contExterno->setCampo($destinodelValor, $valorDelCampo);
                            $ObjCampo->contExterno->setNuevoValorCampo($destinodelValor, $valorDelCampo);
                        }
                    }



                    $abmDatosDet->showTablaInt('noecho', '', '', 'nocant');
                    $Tablatemp2 = $ObjCampo->contExterno->TablaTemporal->datos();
                    $textCab = '';
                    
        			// check if print the header or not;
        			$innerHeaderData   = $this->getHeaderData($ObjCampo->contExterno);
        			$printCab = !$this->isRowEmpty($innerHeaderData['headerRow']);


                    $this->printCabint[$ObjCampo->contExterno->xml] = $printCab;

                    //$n=0;
                    foreach ($ObjCampo->contExterno->tablas[$ObjCampo->contExterno->TablaBase]->campos as $ncam => $campo) {
                        $arrayC[$ncam] = $campo;
                        if ($campo->resize == 'false')
                            $this->noResize[$ncam] = $ncam;
                    }
                    $anchoC = $this->getAnchoTabla($ObjCampo->contExterno, $Tablatemp2);
                    $anchoC = $this->ajustoAnchos($anchoC, $anchoColumna[$nombreCampo]);
                    $nb2 = 0;

                    if ($Tablatemp2) {
                        loger($Tablatemp2);
                        foreach ($Tablatemp2 as $nrow2 => $row2) {
                            $nb2 += $this->getAltoFila($row2, $arrayC, $anchoC, $ObjCampo->contExterno, $this->printCabint[$ObjCampo->contExterno->xml]) ;
                            
                        }

                        if ($this->printCabint[$ObjCampo->contExterno->xml] != false) {
                            $nb2 += 1;
                        }
                        //else $nb2 += 0.5;

                    }

                    $nb = max($nb, $nb2 );

                    //return $nb;
                    //$nb =max($nb, count($Tablatemp2) + 1);
                    // if exists a sum row add another row to total height
                    if (array_sum($abmDatosDet->Suma) > 0) {

                        $nb += 1;
                    }
                }

                if (isset($anchoColumna[$nombreCampo])) {
                    $nb = max($nb, $this->NbLines($anchoColumna[$nombreCampo], trim(str_replace('.', '', $valor))));
                }
            }
        }
        return $nb;
    }
	
	/*
	* Get Header Data
	*  returns an array with columns width and header row
	*/
	function getHeaderData($MisDatos, $anchoColumna=''){
		
        // Calculo en ancho de las columnas para las cabeceras
        if (is_object($MisDatos)) {
            $campos = $MisDatos->camposaMostrar();
            foreach ($campos as $n => $valor) {

                // inicializo totalizadores
                $ObjCampo = $MisDatos->getCampo($valor);
                if (isset($ObjCampo->Oculto) && ($ObjCampo->Oculto || $ObjCampo->Oculto == 'true'))
                    continue;
                if (isset($ObjCampo->print) && $ObjCampo->print == 'false')
                    continue;
                if (isset($ObjCampo->noshow) && $ObjCampo->noshow == 'true')
                    continue;
                if (isset($ObjCampo->colstyle) && strpos('_' . $ObjCampo->colstyle, 'display:none'))
                    continue;
                if (isset($ObjCampo->noEmpty) && $ObjCampo->noEmpty == 'true' && !isset($MisDatos->hasValue[$ObjCampo->NombreCampo])) {
                    continue;
                }

                $valorEt = $ObjCampo->Etiqueta;
                $valorEt = html_entity_decode(strip_tags($valorEt), ENT_QUOTES);

                $ancho = $this->GetStringWidth(trim($valorEt));
                if (!isset($anchoColumna[$valor]))
                    $anchoColumna[$valor] = $ancho;    
                $anchoColumna[$valor] = max($ancho, $anchoColumna[$valor]);

                $cabecera[$valor] = $valorEt;
            }
			$headerData['headerRow'] = $cabecera;
			$headerData['headerWidth'] = $anchoColumna;
			
			
			return $headerData;
			}
	}



    /**
     * Impresion de Tablas
     * @param ContDatos $MisDatos  Datacontainer
     * @param Table $TablaAlt  Alternate Table???
     * @param string $opt  options
     * @param string $anchoTabla  Table Width
     * @param integer $fontsize  FontSize
     * @param integer $posX  FontSize
     * @param bool $printCab  Print header
     */
    function impTabla($MisDatos, $TablaAlt =null, $opt=null, $anchoTabla=null, $fontsize=null, $posX=null, $printCab=true) {
        if (isset($MisDatos->PDFlines) && $MisDatos->PDFlines != '') {
            $opt['lineas'] = $MisDatos->PDFlines;
        }
        if (isset($MisDatos->PDFtitles) && $MisDatos->PDFtitles != '') {
            $opt['titulo'] = $MisDatos->PDFtitles;
        }


        if (isset($MisDatos->PDFfontSize)) {
            $fontsize = $MisDatos->PDFfontSize;
        }
        if ($fontsize != null)
            $this->fontsize = $fontsize;
        $fontsize = $this->fontsize;

        $this->SetFontSize($fontsize );

        $fieldSums = [];


        $this->SetFillColor(255, 255, 255);
        $this->SetDrawColor(192, 192, 192);
//        $this->SetDrawColor(0, 0, 0);

        $this->SetTextColor(0);
        $this->SetFont('','');
        $fila = 0;
        if ($posX == null) {
            $posX = $this->margenX / 2 + $this->offsetX;
        }
        //$posX = 0;
        // Calculo en ancho de las columnas para los datos
        if ($TablaAlt != null)
            $Tablatemp = $TablaAlt;
        else {
            if (is_object($MisDatos))
                $Tablatemp = $MisDatos->TablaTemporal->datos();
        }

        if ($anchoTabla == null && $posX != null) {

            $anchoTabla = $this->anchoPagina - $posX;
        }
        $anchoColumna = $this->getAnchoTabla($MisDatos, $Tablatemp);
        if ($anchoTabla == 'auto') {

            if (is_array($anchoColumna))
                $anchoTabla = array_sum($anchoColumna);
            else
                $anchoTabla = $anchoColumna;
        }



        $break = false;
        $sizeOrig = $this->fontsize;
        $size = $sizeOrig;
        $rownum = -1;
        $Height = 0;

        $this->SetFontSize($size );
        $cantRows = count($Tablatemp);



        if ($posX != null)
            $this->setX($posX);

        if ($Tablatemp != '')
            foreach ($Tablatemp as $orden => $row) {
                $rownum++;
                foreach ($row as $nombreCampo => $valor) {
                    $ObjCampo = $MisDatos->getCampo($nombreCampo);


                    if (!is_object($ObjCampo))
                        continue;

                    // Set Attributes from row values
                    $ObjCampo->setAttributes($row);

                    // Break by Data
                    if (isset($MisDatos->hasBreak) && $MisDatos->hasBreak) {

                        if ($MisDatos->seSuma($nombreCampo))
                            $partialSum[$nombreCampo] = $valor;
                        else
                            $partialSum[$nombreCampo] = ' ';

                        if (isset($ObjCampo->break) && $ObjCampo->break == 'true') {
                            if ($orden != 0 && $oldData[$nombreCampo] != $valor) {
                                $break = true;
                            }
                            $oldData[$nombreCampo] = $valor;
                        }
                    }

                    // Remove Non printables
                    if (isset($ObjCampo->Oculto) && ($ObjCampo->Oculto || $ObjCampo->Oculto == 'true'))
                        continue;
                    if (isset($ObjCampo->noshow) && $ObjCampo->noshow == 'true')
                        continue;
                    if (isset($ObjCampo->colstyle) && strpos('_' . $ObjCampo->colstyle, 'display:none'))
                        continue;
                    if (isset($ObjCampo->noEmpty) && $ObjCampo->noEmpty == 'true' && !isset($MisDatos->hasValue[$ObjCampo->NombreCampo])) {
                        continue;
                    }



                    if (isset($ObjCampo->resize) && $ObjCampo->resize == 'false') {
                        $this->noResize[$ObjCampo->NombreCampo] = $ObjCampo->NombreCampo;
                    }


			        $this->_rowId = $orden;
		    
                    unset($ObjCampo->opcion);
    		                              
                    if ($ObjCampo->isSelect ){
                       //set each datacontainer
                       // this will refresh every data container or every row ALL the time.
                       if ($ObjCampo->helperXml != '') {
                           $xmlReader = new Histrix_XmlReader($MisDatos->dirXmlPrincipal, $ObjCampo->helperXml, true, $MisDatos->xml, $ObjCampo->helperDir,true);
                           $micont = $xmlReader->getContainer();

                           $micont->xml = $helperXml;
                            $ObjCampo->contExterno  = $micont;
                            $deleteInnerContainer = true;
                        }
                        else {

                        }

    				    if ($ObjCampo->showObjTable=="true" || $ObjCampo->showObjTabla == "true"){
    //					$deleteInnerContainer = false;
    //					echo $ObjCampo->NombreCampo."true";
    					}


                        // get new options
                        if (is_object($ObjCampo->contExterno)){
                           $ObjCampo->refreshInnerDataContainer($MisDatos, $row);
                            $ObjCampo->llenoOpciones('false',$orden);

                        }

                        $ObjCampo->customRowName = true;
                    }
                        
			     
                    if ($deleteInnerContainer === true){
                        unset($ObjCampo->contExterno);
                    }
                         
                                       

                    // Si tiene opciones de un combo

                    $valop = (isset($ObjCampo->valop)) ? $ObjCampo->valop : '';
                    if (isset($ObjCampo->opcion) && count($ObjCampo->opcion) > 0 && $ObjCampo->TipoDato != "check" && $valop != 'true') {
                        //$align = ' align="left" ';
                        if (isset($ObjCampo->opcion[$valor])) {
                            $valor = $ObjCampo->opcion[$valor];
                            if (is_array($valor))
                                $valor = current($valor);
                        }
                    }
    	
    	            $valor = $this->ReplaceHTML($valor);
                    $valorMed = $valor;


	            $FT = 'FieldType_'.$ObjCampo->TipoDato;
                if ($ObjCampo->TipoDato != ''){
    	            if (is_file(dirname(__FILE__).'/../FieldType/'.$FT.'.php')){
        	            /*
        	            $alignCons = constant($FT.'::ALIGN');
            		    if ($alignCons != 'left')
                    		$align   = ' align="'.$alignCons.'" ';

                   
        	            $arrayAtributos['dir']     = constant($FT.'::DIR');
        	            */

        	             // Add custom Parameters
            		     //   $modif['custom'] = $FT::customCellParameters();


                	    if (method_exists($FT ,'printValue')){
                    		$valor = $FT::printValue($valor, $ObjCampo);
                    		$valorMed = $valor;
                	    }

        	        }
            		else  {

        	            loger('falta: '.dirname(__FILE__).'/../FieldType/'.$FT.'.php', 'fieldtypes_ui.log');
            		}
                }


                    switch ($ObjCampo->TipoDato) {
                        case "integer" :
                            break;
                        case "decimal" :
                        case "numeric" :
                        case "custom_numeric" :
                            $precision = ($ObjCampo->numberPrecision != '')? $ObjCampo->numberPrecision:2;
                            if (is_numeric($valor))
                                $valorMed = number_format($valor, $precision, ',', '.');
                            break;
                        case "date" :
                            if ($valor == '0000-00-00')
                                $valorMed = '';
                            else
                                $valorMed = 'XX/XX/XXXX';
                            break;
                        case "file":
                            // add files to file array
                            // used for  attachs in emails
                            $file = new Archivo($valor, '', '' , $objCampo);
                            if ($valor != '') {
                                $link = $ObjCampo->url;
                                if ($MisDatos->path != '') {
                                    $filePath = '../database/' . $_SESSION['datapath'] . 'xml/' . $MisDatos->path;
                                    $link = $filePath . '/';
                                }
                                
                                // Add Object Path
                                if ($ObjCampo->path != '') {
                            	
                            	    if (isset($row[$ObjCampo->path]))
                                        $Objurl = $row[$ObjCampo->path];
                                	else 
	                                    $Objurl = $ObjCampo->path;
                                        $link .= $Objurl . '/';
                                }
                                $link .=$file->link;

                                if ($ObjCampo->watermark != '') {
                                    $fileName = $row[$ObjCampo->watermark];
                                }

                                if (is_file($link))
                                    $this->attachedFiles[$link] = $fileName;
                            
                            }
                            break;
                    }

                    if (isset($ObjCampo->print) && $ObjCampo->print == 'false')
                        continue;


                    if (isset($ObjCampo->prefijo) && $ObjCampo->prefijo != '') {
                        // hack to add width to string
                        $valorMed = $ObjCampo->prefijo . 'XX' . $valor;
                    }


                    /* FORMATEO EXPLICITO */


		    //$valor = $ObjCampo->getFormatedValue($valor);
		    
                    if (isset($ObjCampo->Formato)) {
                        $formatoCampo = $ObjCampo->Formato;

                        if ($formatoCampo != '') {
                            // las fechas se formatean diferente
                            if ($ObjCampo->TipoDato == 'date' || $ObjCampo->TipoDato == 'time') {
                                //loger($valor);
                                //$valor = date($formatoCampo, strtotime($ObjCampo->valor));
                                $valor = date($formatoCampo, strtotime($valor));
                            } else {
                                $valor = sprintf($formatoCampo, $valor);
                            }
                        }
                    } 
//                    $valorMed = $valor;

                    $ancho = 0;

                    $ancho = $this->GetStringWidth(trim($valorMed));


                    switch ($ObjCampo->TipoDato) {
                        case "grafico" :
                            $ancho = 10;
                            break;
                    }

                    // Para el caso de tablas en tablas (picture in picture je)

                    $arrayCampos[$nombreCampo] = $ObjCampo;

                    $anchoColumna[$nombreCampo] = max($ancho, $anchoColumna[$nombreCampo]);

                    $Tablatemp[$rownum][$nombreCampo] = $valor;
                }

                // Break data
                if (isset($MisDatos->hasBreak) && $MisDatos->hasBreak) {

                    if ($break) {
                        $currentRow = $Tablatemp[$rownum];

                        foreach ($row as $nombreCampo => $valorCampo) {
                            $value = '';
                            if ($MisDatos->seSuma($nombreCampo))
                                $value = $breakTotal[$nombreCampo];
                            $Tablatemp[$rownum][$nombreCampo] = $value;
                            $Metadata[$rownum] = 'sumrow';
                            $break = false;
                        }
                        $rownum++;
                        $Tablatemp[$rownum] = $currentRow;

                        unset($breakTotal);
                    }

                    foreach ($partialSum as $n => $val) {
                        $breakTotal[$n] += $val;
                    }
                    unset($partialSum);
                }


                // Break data
                if (isset($MisDatos->hasBreak) && $MisDatos->hasBreak) {

                    if ($orden + 1 == $cantRows) {
                        $rownum++;
                        foreach ($row as $nombreCampo => $valor) {
                            $value = '';
                            if ($MisDatos->seSuma($nombreCampo))
                                $value = $breakTotal[$nombreCampo];
                            $Tablatemp[$rownum][$nombreCampo] = $value;
                            $Metadata[$rownum] = 'sumrow';
                            $break = false;
                        }
                    }
                }
            }

        //$anchoColumna = $this->getAnchoTabla($MisDatos, $Tablatemp);
        // Calculo en ancho de las columnas para las cabeceras
        
        if (is_object($MisDatos)) {
  
			$headerData   = $this->getHeaderData($MisDatos, $anchoColumna);
			
			$cabecera     = $headerData['headerRow'];
			$anchoColumna = $headerData['headerWidth'];
			
            // Calculo los Totales
            $campos = $MisDatos->camposaMostrar();
            $this->SetFont('');
            foreach ($campos as $n => $nomcampo) {
                $ObjCampo = $MisDatos->getCampo($nomcampo);


                if (isset($ObjCampo->Oculto) && ($ObjCampo->Oculto || $ObjCampo->Oculto == 'true'))
                    continue;
                if (isset($ObjCampo->print) && $ObjCampo->print == 'false')
                    continue;
                if (isset($ObjCampo->noshow) && $ObjCampo->noshow == 'true')
                    continue;

                if (isset($ObjCampo->colstyle) && strpos('_' . $ObjCampo->colstyle, 'display:none'))
                    continue;

                if (isset($ObjCampo->noEmpty) && $ObjCampo->noEmpty == 'true' && !isset($MisDatos->hasValue[$ObjCampo->NombreCampo])) {
                    continue;
                }



                if ($MisDatos->seSuma($nomcampo))
                    $totales[$nomcampo] = $ObjCampo->Suma;
                else
                    $totales[$nomcampo] = '';


                if (trim($totales[$nomcampo]) == '')
                    continue;


                $ancho = $this->GetStringWidth(trim($totales[$nomcampo]));
                $anchoColumna[$nomcampo] = max($ancho, $anchoColumna[$nomcampo]);
            }
            unset($anchoColumna['']);

            $anchoColumna = $this->ajustoAnchos($anchoColumna, $anchoTabla);

            $this->Columnwidths = $anchoColumna;

            // Imprimo la cabecera
            if ($opt != '') {
                if ($opt['titulo'] == false ||
                        $opt['titulo'] == 'false')
                    $printCab = false;
            }
            ;

            $this->setX($posX);

            if ($printCab === true && $Tablatemp != '') {
                $this->impCabTabla($cabecera, $anchoColumna, null, $fontsize);
            }

            $campos = $MisDatos->camposaMostrar();

            if ($Tablatemp != '')
                foreach ($Tablatemp as $orden => $row) {

                    $nb = 0;
                    $cols = count($row);
                    $clasefila = '';

                    /* Fix round numbers REDO ALL THIS TO AVOID DUPLICATION */
                    foreach ($campos as $n => $nombreCampo) {
                        $valor = null;
                        if (isset($row[$nombreCampo]))
                            $valor = $row[$nombreCampo];

                        if (isset($arrayCampos[$nombreCampo])) {
                            $ObjCampo = $arrayCampos[$nombreCampo];


	                // Set Attributes from row values
                        $ObjCampo->setAttributes($row);


                            switch ($ObjCampo->TipoDato) {
                                case "decimal" :
                                case "numeric" :
                                case "custom_numeric" :

                                    $precision = ($ObjCampo->numberPrecision != '')? $ObjCampo->numberPrecision:2;
                                    $align = 'R';
                                    if ($formatoCampo == '' && is_numeric($valor)){
                                	
                                        $valor = number_format($valor, $precision, ',', '.');
                                    }
                                        
                                    break;
                            }
                        }

                        $row2[$nombreCampo] = $valor;
                    }

                    // Clase Fila
                    foreach ($row as $nombreCampo => $valor) {
                        $ObjCampo = $MisDatos->getCampo($nombreCampo);
                        $valor = $row[$nombreCampo];
                        if (isset($ObjCampo->clasefila) && $ObjCampo->clasefila != '') {
                            $clasefila = $valor;
                        }
                    }

                    $nb = $this->getAltoFila($row2, $arrayCampos, $anchoColumna, $MisDatos, $printCab);

                    $h = ($this->fontsize / 2) * $nb;

                    $Height += $h;

                    $salto = false;

                    if (isset($MisDatos->PDFpageBreak) && $MisDatos->PDFpageBreak == 'false') {
                        // do not page break
                    }
                    else {
                        $salto = $this->checkPageBreak($h, $fieldSums, $arrayCampos);

                    }
                    $this->setX($posX);


                    if (($salto)) {
                        $this->Ln(5);
                        $this->setX($posX);
                        if ($printCab == true)
                            $this->impCabTabla($cabecera, $anchoColumna, null, $fontsize);
                    }

                    foreach ($campos as $n => $nombreCampo) {
                        $valor = null;
                        if (isset($row[$nombreCampo]))
                            $valor = $row[$nombreCampo];

                        if (!isset($arrayCampos[$nombreCampo]))
                            continue;

                        $ObjCampo = $arrayCampos[$nombreCampo];


            			// remove value if repeated
            			if ($ObjCampo->repeat=='false'){
            			    if ($orden > 0){                                            
            				if ($row[$nombreCampo] == $Tablatemp[$orden - 1][$nombreCampo])
            					$valor = '';
            			    }
            			}



                        if (isset($ObjCampo->Oculto) && ($ObjCampo->Oculto || $ObjCampo->Oculto == 'true'))
                            continue;
                        if (isset($ObjCampo->print) && $ObjCampo->print == 'false')
                            continue;
                        if (isset($ObjCampo->noshow) && $ObjCampo->noshow == 'true')
                            continue;

                        if (isset($ObjCampo->colstyle) && strpos('_' . $ObjCampo->colstyle, 'display:none'))
                            continue;
                        if (isset($ObjCampo->noEmpty) && $ObjCampo->noEmpty == 'true' && !isset($MisDatos->hasValue[$ObjCampo->NombreCampo])) {
                            continue;
                        }


                        if (isset($ObjCampo->PDFsum) && $ObjCampo->PDFsum == 'true') {
                            $rawval = $valor;
                            if (isset($fieldSums[$ObjCampo->NombreCampo])) {
                                $fieldSums[$ObjCampo->NombreCampo] += (float) $rawval;
                            } else {
                                $fieldSums[$ObjCampo->NombreCampo] = (float) $rawval;
                            }
                        } else {
                            $fieldSums[$ObjCampo->NombreCampo] = '';
                        }

                        //if ($this->Datos->seAcumula($nomcampo) && $this->norepeat != true) {
                        if (isset($MisDatos->acumulaCampo[$ObjCampo->NombreCampo])) {


                            $this->PartialSum[$ObjCampo->NombreCampo] += (float) str_replace(',', '', $valor);

                            $valor = number_format($this->PartialSum[$ObjCampo->NombreCampo], 2, '.', ',');

                        }


                        if ($anchoColumna[$nombreCampo] == '')
                            continue;

                        $x = $this->GetX();
                        $y = $this->GetY();
                        if (isset($minXTabla))
                            $minXTabla = min($minXTabla, $x);
                        else
                            $minXTabla = $x;

                        $align = 'J';
                        $estilo = 'D';
                        $img = '';
                        $fill = 0;

                        $estiloFont = '';
                        $borde = '0';
                        $sizeOrig = $this->fontsize;
                        $size = $sizeOrig;
                        $this->SetFontSize($size);



                        $formatoCampo = (isset($ObjCampo->Formato)) ? $ObjCampo->Formato : '';

                        $this->SetTextColor(0);

                        if (isset($ObjCampo->esClave) && $ObjCampo->esClave == 'true') {
                            $this->SetTextColor(250, 0, 0);
                        }
                        if (isset($ObjCampo->PDFcolor) && $ObjCampo->PDFcolor != '') {
                            $textrgb = hex_to_rgb($ObjCampo->PDFcolor);
                            $this->SetTextColor($textrgb['red'], $textrgb['green'], $textrgb['blue']);
                        }

                        if (isset($ObjCampo->PDFbackgroundColor) && $ObjCampo->PDFbackgroundColor != '') {
                            $textrgb2 = hex_to_rgb($ObjCampo->PDFbackgroundColor);
                            $this->SetFillColor($textrgb2['red'], $textrgb2['green'], $textrgb2['blue']);
                            $estilo = 'DF';

                        }

                        if (isset($Metadata) && $Metadata[$orden] == 'sumrow') {
                            $estilo = 'DF';
                            $this->SetFillColor(230, 255, 214);
                            $this->SetFont($defaultFont, 'B', $size);
                        }
                        if ($clasefila != '') {
                            $estilo = 'DF';
                            $this->SetFillColor(255, 233, 153);
                        }

                        // TODO Remove THIS FROM INSIDE FOREACH
                        $this->SetDrawColor(192, 192, 192);

                        if (isset($MisDatos->PDFlineColor) && $MisDatos->PDFlineColor != '') {
                            $color = explode(',', $MisDatos->PDFlineColor);
                            $this->SetDrawColor($color[0], $color[1], $color[2]);
                        }

                        $lineas = true;
                        if ($opt != '') {
                            if ($opt['lineas'] == false || $opt['lineas'] == 'false') {
                                $lineas = false;
                                //	$this->SetDrawColor(255, 255, 255); //por las dudas
                            }
                        }

                        if (isset($opt['tree']) && $opt['tree'] == true) {
                            $lineas = true;
                            $this->SetDrawColor(192, 192, 192);
                        }

                        switch ($ObjCampo->TipoDato) {
                            case "integer" :
                                $align = 'R';
                                break;
                            case "decimal" :
                            case "numeric" :
                            case "custom_numeric" :

                                $precision = ($ObjCampo->numberPrecision != '')? $ObjCampo->numberPrecision:2;
                                $align = 'R';
                                if ($formatoCampo == '' && is_numeric($valor))
                                    $valor = number_format($valor, $precision, ',', '.');
                                break;
                            case "time" :
                                $align = 'C';
			    break;

                            case "date" :
                                $align = 'C';
                                if ($valor == '0000-00-00')
                                    $valor = '';
                                if ($valor != '') {
                                    if ($formatoCampo != '') {
                                        //loger($valor.' -- '.$formatoCampo);
                                        //$valor = date($formatoCampo, strtotime($valor));
                                    }
                                    else
                                        $valor = date("d/m/Y", strtotime($valor));
                                }
                                break;
                            case "isbn" :
                                $valor = isbn102isbn13($valor);
                                break;
                            case "check" :
                                $this->SetDrawColor(0, 0, 0);
                                $this->CheckBox(($x + 0.5), ($y + 0.5), $valor);
                                $this->SetDrawColor(192, 192, 192);
                                $img = 'true';
                                break;
                            case "simpleditor" :
                            case "editor":
                                $align = 'J';
                                $valor = str_ireplace('<br>', "\n", $valor);
                                $valor = strip_tags($valor);

                                break;

                            case "email" :
                                break;
                            case "grafico" :
                                $align = 'C';
                                if (isset($ObjCampo->max)) {
                                    $max = $ObjCampo->max;
                                    $min = 0;
                                    $showval = 1;
                                } else {
                                    $max = $MisDatos->TablaTemporal->getMax($ObjCampo->NombreCampo);
                                    $min = $MisDatos->TablaTemporal->getMin($ObjCampo->NombreCampo);
                                }
//                                $max = $MisDatos->TablaTemporal->getMax($nombreCampo);
//                                $min = $MisDatos->TablaTemporal->getMin($nombreCampo);
                                $clase = 'grHoriz';
                                $anchoIMG = 100;
                                $altoIMG = 20;
                                $uid = uniqid();

                                if ($ObjCampo->grafico == 'barcode')
                                    $clase = 'barcode';
                                $img = new Graficar($clase, $anchoIMG, $altoIMG, null, 'jpg');
                                $estilo = 'D';
                                switch ($clase) {
                                    case "grHoriz":
                                        $c1 = (isset($ObjCampo->color1)) ? hex_to_rgb($ObjCampo->color1) : array('red' => 0, 'green' => 255, 'blue' => 0); //verde;
                                        $c2 = (isset($ObjCampo->color2)) ? hex_to_rgb($ObjCampo->color2) : array('red' => 255, 'green' => 0, 'blue' => 0); //rojo;

                                        $img->crearImagen();
                                        $img->grHoriz($valor, $min, $max, $c1, $c2, $showval);

                                        

                                        break;
                                    case "barcode":
                                        if ($valor != '') {
                                            //		$img->crearImagen();
                                            $img->barcode($valor, $altoIMG, 1.5, $uid, $ObjCampo->Parametro['encode']);
                                        }
                                        break;
                                }
                                $imagen = '../database/' . $_SESSION['datapath'] . 'tmp/' . $uid . '.jpg';
                                if ($img->imagen != '')
                                    imagejpeg($img->imagen, $imagen);

                                if ($valor != ''){
				    if (is_file($imagen))
        	                            $this->Image($imagen, $x, $y, $anchoColumna[$nombreCampo] - 5, $h);
					    
				}
                                if ($img->imagen)
                                    imagedestroy($img->imagen);
                                if ($valor != '')
                                    unlink($imagen);
                                break;


                            case "file" :
                                //$file = new Archivo($valor, $this->Datos->path, '../archivos/');
                                $file = new Archivo($valor, '', '', $ObjCampo);
                                if ($valor != '') {
                                    $link = $ObjCampo->url;

                                    // Path for uploading files
                                    //$link .= $MisDatos->path.'/';
                                    if ($MisDatos->path != '') {
                                        $filePath = '../database/' . $_SESSION['datapath'] . 'xml/' . $MisDatos->path;
                                        $link = $filePath . '/';
                                    }

                                    // Add Object Path
                                    if ($ObjCampo->path != '') {
                                        //	$filePath= '../xml/'.$_SESSION['db'].'/'.$MisDatos->path;
                                        //	$link = $filePath;
                                        //$pathField= $MisDatos->getCampo($ObjCampo->path);

					if (isset($row[$ObjCampo->path]))
                                            $Objurl = $row[$ObjCampo->path];
                                        else 
                                    	    $Objurl = $ObjCampo->path;

                                        /*
                                         * if ($pathField->valor !='')
                                          $Objurl = $pathField->valor;
                                          else
                                          $Objurl = $pathField->ultimo;
                                         * */


                                        //loger(print_r($pathField, true));
                                        //     echo '--'.$Objurl;
                                        $link .= $Objurl . '/';
                                        //echo
                                    }
                                    $link .=$file->link;

                                    if (is_file($link))
                                        $this->AttachedFiles[$link] = $link;

                                    if (($file->imagen || $file->preview == true ) && $file->svg != true) {
                                        //echo $link;
                                        if (is_file($link)) {

                                            $img = $link;
                                            $Imgsize = ($ObjCampo->imageWidth != '') ? $ObjCampo->imageWidth : 10;
                                            $imageResize = (isset($ObjCampo->resample) && $ObjCampo->resample == 'false') ? false : true;

                                            /* Resize Image so the PDf is lighter */
                                            $source_pic = 'http://../principal';
                                            $https = ($_SERVER['HTTPS'] == 'on') ? 'https' : 'http';
                                            $server = $https . '://' . $_SERVER['SERVER_NAME'];

                                            $source_pic      = $server . dirname($_SERVER['PHP_SELF']) . '/thumb.php?url=' . $link . '&ancho=' . $Imgsize;
                                            $destination_pic = $file->toImage($link);

                                            if ($width == 0)
                                                $width = 10;
                                            if ($height == 0)
                                                $height = 10;

                                            $max_width = 800;
                                            $max_height = 600;

                                            $width = $file->width;
                                            $height = $file->height;

                                            // Just DownSize Images
                                            if ($imageResize) {
                                                $width = ($file->width < $max_width ) ? $file->width : $max_width;
                                                $height = ($file->height < $max_height) ? $file->height : $max_height;
                                            }

                                            $destination_pic = $file->toImage($link, $width, $height);

                                            $imageHeightMargin = 1;

                                            $tmpfw = $this->fw;
                                            $tmpfh = $this->fh;

                                            if ($this->DefOrientation == 'L') {
                                                $this->fw = $tmpfh;
                                                $this->fh = $tmpfw;
                                            }
                                            //echo $this->fw;
                                            if ($Imgsize < $anchoColumna[$nombreCampo])
                                                $centerMargin = ( $anchoColumna[$nombreCampo] - $Imgsize ) / 2;
                                            $imgaMargin = $ObjCampo->imageMargin;
//	                                echo  $centerMargin;
					    if (is_file($destination_pic)){
            	                                $sizesarray = $this->Image($destination_pic, $x + $centerMargin + $imageMargin, $y, $Imgsize, null, 'jpg');
					    }

                                            $this->fw = $tmpfw;
                                            $this->fh = $tmpfh;

                                            //        unlink($destination_pic);

                                            $h = $sizesarray['HEIGHT'] + $imageHeightMargin * 2;

                                            $nb = ($this->fontsize / 2) / $h;
                                        }
                                        else {
                                    	// file not found
                                    //	echo $link;
                                        
                                        }
                                    } else {
                                        
                                    }
                                }
                                break;
                        }

	    	            $valor = $this->ReplaceHTML($valor);

                        if (isset($ObjCampo->noZero) && $ObjCampo->noZero == 'true') {
                            if ($valor === 0 || $valor === '0' || $valor === '0,00')
                                $valor = '';
                        }

                        if (isset($ObjCampo->prefijo) && $ObjCampo->prefijo != '')
                            $valor = $ObjCampo->prefijo . $valor;

                        if ($lineas)
                            $this->Rect($x, $y, $anchoColumna[$nombreCampo], $h, $estilo);

                        if ($totales[$nombreCampo] != '') {
                            $estilo = 'DF';
                            $this->SetFillColor(230, 255, 214);
                        }

                        /* 		if ($img !='') $estilo = 'D';

                          $this->SetDrawColor(192, 192, 192);
                          $this->Rect($x, $y, $anchoColumna[$nombreCampo], $h, $estilo); */

                        // Para el caso de tablas en tablas (picture in picture je)

                        if ( isset($ObjCampo->contExterno)  
                        && isset($ObjCampo->esTabla) && $ObjCampo->esTabla 
                        ) 
                        {
                        

                            if ($ObjCampo->paring != '') {

                                foreach ($ObjCampo->paring as $destinodelValor => $origendelValor) {
                                
                                    $valorDelCampo = $row[$origendelValor['valor']];
                                    if ($valorDelCampo == '')
                                        $valorDelCampo = '0';
                                    if ($MisDatos->getCampo($origendelValor['valor'])->TipoDato == 'varchar')
                                        $comillas = '"';
                                    $operador = '';
                                    $operador = $origendelValor['operador'];
                                    if ($operador == '')
                                        $operador = '=';
                                    $reemplazo = '';
                                    $reemplazo = $origendelValor['reemplazo'];
                                    if ($reemplazo != 'false')
                                        $reemplazo = 'reemplazo';
                                    else
                                        $reemplazo = '';

                                    $ObjCampo->contExterno->addCondicion($destinodelValor, $operador, $comillas . $valorDelCampo . $comillas, 'and', $reemplazo, false);

                                    $ObjCampo->contExterno->setCampo($destinodelValor, $valorDelCampo);
                                    $ObjCampo->contExterno->setNuevoValorCampo($destinodelValor, $valorDelCampo);
                                }
                            }
                            $UI = 'UI_' . str_replace('-', '', $ObjCampo->contExterno->tipo);
			                 unset($abmDatosDet);

                            $ObjCampo->contExterno->unserializeParent= 'false';

                            $abmDatosDet = new $UI($ObjCampo->contExterno);

            				// check if print the header or not;
        		    	    $innerHeaderData   = $this->getHeaderData($ObjCampo->contExterno);
        			        $printCab = !$this->isRowEmpty($innerHeaderData['headerRow']);

                            $abmDatosDet->showTablaInt('noecho', '', '', 'nocant');
                            $ccc++;
                            $fontSize = $this->fontsize;

                            if ($this->printCabint[$ObjCampo->contExterno->xml] != 'false')
                                $this->printCabint[$ObjCampo->contExterno->xml] = $printCab;

                            if ($ObjCampo->showValor == "true" && $ObjCampo->linkint == '') {

                                $valor = utf8_decode($valor);
                                $heightRel = 2;
                                $lineH = ($this->fontsize / $heightRel);
                                $this->MultiCell($anchoColumna[$nombreCampo], $lineH, $valor, $borde, $align, $fill);
                                $h += ( $this->fontsize / 2);
                            }
                                               
                            $this->impTabla($ObjCampo->contExterno, null, null, $anchoColumna[$nombreCampo], $fontsize, $x, $this->printCabint[$ObjCampo->contExterno->xml]);


                            $this->printCabint[$ObjCampo->contExterno->xml] = 'false';

                            $this->fontsize = $fontSize;
                        } else {

                            $valor = utf8_decode($valor);
                            if ($img == '') {
                                $heightRel = 2;
                                if (isset($MisDatos->PDFheigthRel) && $MisDatos->PDFheigthRel != '')
                                    $heightRel = $MisDatos->PDFheigthRel;
                                $lineH = ($this->fontsize / $heightRel);

                                $this->MultiCell($anchoColumna[$nombreCampo], $lineH, $valor, $borde, $align, $fill);
                            }
                        }
                        unset($img);
                        $this->SetFont($this->defaultFont, $estiloFont, $size);
                        $this->SetFont('');
                        //Put the position to the right of the cell

                        $this->SetXY($x + $anchoColumna[$nombreCampo], $y);
                        $size = $sizeOrig;
                        $this->SetFontSize($size);
                    }
                    $this->maxY = max($this->maxY, $this->GetY());

                    $opttree = (isset($opt['tree'])) ? $opt['tree'] : '';
                    if ($opttree != true)
                        $this->Ln($h);

                    if ($posX != null)
                        $this->setX($posX);
                }


            $this->impCabTabla($totales, $anchoColumna, $arrayCampos, $this->fontsize);

            // Cantidad de registros
           //
           //  $this->Ln(1);
            /*
             * remove count elements
             *      
             */
             if (count($Tablatemp) >= 1 && $MisDatos->showCantidad == 'true') {
              $this->SetXY(10,$y+5);
              $this->SetFont($this->defaultFont, $estiloFont, 7);
              $heightRel = 2;
              if (isset($MisDatos->PDFheigthRel))
              $heightRel = $MisDatos->PDFheigthRel;
              $lineH = ($this->fontsize / $heightRel);
              $this->MultiCell(0, $lineH, 'Registros: '.count($Tablatemp), $borde, null, $fill);
             } 
             
            $this->SetTextColor(0);
            unset($this->noResize);
        }
        return $Height;
    }

    function impAbm($MisDatos) {

        //$rs = $MisDatos->resultSet;

        $campos = count($MisDatos->camposaMostrar());
        $datos = null;
        $i = 0;
        $fila = 0;
        $col = 0;

        $nombres = $MisDatos->nomCampos();


        while ($i < $campos) {
            $nom = $nombres[$i];

            $ObjNombre = $MisDatos->getCampo($nom);

            // No muestro los campos Ocultos o con posicion absoluta
            if (isset($ObjNombre->print) && $ObjNombre->print == 'false') {
                $i++;
                continue;
            }
            if ($ObjNombre->esOculto()) {
                $i++;
                continue;
            }
            // ignore fixed position elements
            if ($ObjNombre->posX != '') {
                $pdfNolabel = (isset($ObjNombre->PDFnolabel)) ? $ObjNombre->PDFnolabel : '';
                $this->labelXY($ObjNombre, $pdfNolabel, $MisDatos);
                $i++;
                continue;
            }
            if ($ObjNombre->posY == -1)
                $posY = $this->maxY;
            if (substr($ObjNombre->posY, 0, 1) == '+')
                $posY = $this->maxY + $ObjNombre->posY;
            $this->SetY($posY);

            if ($ObjNombre->modpos != 'nobr') {
                $col = 0;
                $fila++;
            }


            if ($ObjNombre->PDFnolabel != 'true') {
                $datos['datos'][$fila][$col] = (string) utf8_decode($ObjNombre->Etiqueta);
                $datos['atributos'][$fila][$col][] = 'label';
                if ($ObjNombre->PDFsize != '') {
                    $datos['atributos'][$fila][$col][] = 'size=' . $ObjNombre->PDFsize; // Etiqueta
                    $this->SetFontSize($objCampo->PDFsize);
                }
                $col++;
            }
            $valor = $ObjNombre->getValor();
            //echo $ObjNombre->valor;
            $precision = ($ObjNombre->numberPrecision != '')? $ObjNombre->numberPrecision:2;
            if ($ObjNombre->TipoDato == 'numeric' ||
                    $ObjNombre->TipoDato == 'custom_numeric'){
                if ($valor == '') $valor = 0;
                $valor = number_format($valor, $precision, ',', '.');
            }

            if ($ObjNombre->TipoDato == 'check') {
                if ($valor == 0)
                    $datos['atributos'][$fila][$col][] = 'nochecked';
                else
                    $datos['atributos'][$fila][$col][] = 'checked';
            }

            if ($ObjNombre->opcion != '')
                $valor = $ObjNombre->opcion[$valor];
            if (is_array($valor))
                $valor = current($valor);

            if ($ObjNombre->TipoDato == 'date' && $valor != '')
                $valor = date("d/m/Y", strtotime($valor));

            if ($ObjNombre->aletras == true) {
                $valor = NumeroALetras($ObjNombre->valor);
            }
            if ($ObjNombre->prefijo != '')
                $valor = $ObjNombre->prefijo . $valor;

            $datos['datos'][$fila][$col] = utf8_decode($valor);
            if ($ObjNombre->PDFsize != '') {
                $datos['atributos'][$fila][$col][] = 'size=' . $ObjNombre->PDFsize; // Etiqueta

                $this->SetFontSize($objCampo->PDFsize);
            }
            $datos['atributos'][$fila][$col][] = 'B';
            $datos['atributos'][$fila][$col][] = 'I';
            $datos['atributos'][$fila][$col][] = 'bordeinf';

            $col++;
            $i++;
        }
        return $datos;
    }

    function impFiltros($MisDatos) {
        $halfFont = $this->fontsize / 2;
        $miny = $this->GetY();
        $maxx = 0;

        if (isset($MisDatos->filtroPrincipal)) {

            $campoFiltro = $MisDatos->campofiltroPrincipal;
            $objCampoContFiltro = $MisDatos->filtroPrincipal->getCampo($MisDatos->campofiltroPrincipal);
            $objCampoFiltro = $MisDatos->getCampo($MisDatos->campofiltroPrincipal);
            $opcionesCampo = $objCampoFiltro->opcion[$objCampoFiltro->getValor()];
            $valorCampoFiltro = $opcionesCampo;
            $labelCampoFiltro = utf8_decode($MisDatos->filtroPrincipal->getTitulo());
	    
            $this->SetDrawColor(192, 192, 192);
            $this->SetFillColor(230, 230, 230);
            $align = 'L';
            $borde = 'B';
            $estilo = 'DF';
            $xold = $this->GetX();
            $yold = $this->GetY();
            $miny = min($yold, $miny);

            $this->SetFont($this->defaultFont, 'B');
            $anchoLabel = $this->GetStringWidth(trim($labelCampoFiltro)) + 5;

            $this->Rect($xold, $yold, $anchoLabel, ($halfFont), $estilo);
            $this->MultiCell($anchoLabel, ($halfFont), $labelCampoFiltro, $borde, $align, $estilo);

            $this->SetXY($this->getX() + $anchoLabel, $yold);
            $this->SetFont($this->defaultFont);


            $anchoValor = $this->GetStringWidth(trim($valorCampoFiltro)) + 5;
            $borde = '';
            $this->MultiCell($anchoValor, ($halfFont), $valorCampoFiltro, $borde, $align, $estilo);

            $maxx = max($maxx, $anchoLabel + $anchoValor);
        }

        if (isset($MisDatos->filtros))
            foreach ($MisDatos->filtros as $nomfiltro => $objFiltro) {
                $col = 0;
                $CampoFiltro = $MisDatos->getCampo($objFiltro->campo);
                /* print filter even if not selected
                if (isset($CampoFiltro->print) && $CampoFiltro->print == "false")
                    continue;
                    */
                if (isset($CampoFiltro->PDFsize) && $CampoFiltro->PDFsize != '') {
                    $this->SetFontSize($CampoFiltro->PDFsize);
                    $this->fontsize = $CampoFiltro->PDFsize;
                }
                if ($CampoFiltro == null)
                    continue;

                $align = 'L';
                $borde = 'B';
                $estilo = 'DF';
                $soloborde = true;
                $this->SetDrawColor(192, 192, 192);
                $this->SetFillColor(230, 230, 230);
                $this->SetFont($this->defaultFont, 'B');

                $yold = $this->getY();
                $miny = min($yold, $miny);
                $xold = $this->getX();

                $this->SetFont($this->defaultFont);
                $valor = $objFiltro->valor;

                if ($valor == '')
                    $valor = $CampoFiltro->valor;

                if (isset($CampoFiltro->opcion)) {
                    $valor = $CampoFiltro->opcion[$valor];
                    //print_r(  $valor);
                    if ($valor == '') {
                        $valor = current($CampoFiltro->opcion);
                    }
                    if (is_array($valor))
                        $valor = current($valor);
                }

		if ($valor == '') continue;

                if ($objFiltro->label != '') {
                    $label = utf8_decode($objFiltro->label);
                    $anchoLabel = $this->GetStringWidth(trim($label)) + 5;
                    $this->Rect($xold, $yold, $anchoLabel, ($halfFont), $estilo);
                    $this->MultiCell($anchoLabel, ($halfFont), $label, $borde, $align, $estilo);
                    $this->SetXY($this->getX() + $anchoLabel, $yold);
                }

		
                $valor = utf8_decode($valor);
                $anchoValor = $this->GetStringWidth(trim($valor)) + 5;
                $borde = '';
                $this->MultiCell($anchoValor, ($halfFont), $valor, $borde, $align, $estilo);
                $maxx = max($maxx, $anchoLabel, $anchoValor, $anchoLabel + $anchoValor);
            }
        // Recuadro con filtros
        $margen = 2;
        $this->RoundedRect($xold - $margen / 2, $miny - $margen / 2, $maxx + $margen, $this->getY() - $miny + $margen, 2);
    }

    function setAnchoCol($datos) {
        $w = null;
        $maxsize = 1;
        if (isset($datos['datos']))
            foreach ($datos['datos'] as $nfila => $fila) {
                foreach ($fila as $ncol => $valor) {

                    if (isset($datos['atributos'])) {
                        $attrib = strtolower($datos['atributos'][1][$nfila][$ncol]);
                        $size = '';

                        if (strpos($attrib, 'size=') !== false) {

                            parse_str($attrib);

                            $maxsize = ($size >= $maxsize) ? $size : $maxsize;
                            $this->SetFontSize($maxsize);
                        }
                    }
                    $valor = $this->ReplaceHTML($valor);

                    $valor = 'XX' . $valor . 'XX';
                    if (isset($w[$ncol]))
                        $w[$ncol] = max($this->GetStringWidth($valor), $w[$ncol]);
                    else
                        $w[$ncol] = $this->GetStringWidth($valor);
                }
            }

        return $w;
    }

    function WriteHTML2($html, $posX =5, $customWidth) {
        //HTML parser
        //   $html = str_replace("\n", ' ', $html);

        $html = strip_tags($html, '<p><a><br><u><b><h1><h2><h3><strong><page>');

        $a = preg_split('/<(.*)>/U', $html, -1, PREG_SPLIT_DELIM_CAPTURE);
        foreach ($a as $i => $e) {
            if ($i % 2 == 0) {
                //Text
                if ($this->HREF)
                    $this->PutLink($this->HREF, $e);
                else {

                    $this->Write(5, $e, $posX);
                    //$this->MultiCell(0, 5, $e, 0, 'J');
                }
            } else {
                //Tag
                if ($e { 0 } == '/')
                    $this->CloseTag(strtoupper(substr($e, 1)));
                else {
                    //Extract attributes
                    $a2 = explode(' ', $e);
                    $tag = strtoupper(array_shift($a2));
                    $attr = array();
                    foreach ($a2 as $v) {
//                        if (ereg('^([^=]*)=["\']?([^"\']*)["\']?$', $v, $a3))
                        if (preg_match('/^([^=]*)=["\']?([^"\']*)["\']?$/', $v, $a3))
                            $attr[strtoupper($a3[1])] = $a3[2];
                    }
                    $this->OpenTag($tag, $attr, $posX);
                }
            }
        }
    }

    function convertTag($tag) {
        $tags['STRONG'] = 'B';
        $result = (isset($tags[$tag])) ? $tags[$tag] : $tag;
        return $result;
    }

    function OpenTag($tag, $attr, $posX) {

        $tag = $this->convertTag($tag);

        //Opening tag
        if ($tag == 'B' or $tag == 'I' or $tag == 'U')
            $this->SetStyle($tag, true);
        if ($tag == 'H1' or $tag == 'H2' or $tag == 'H3' or $tag == 'STRONG')
            $this->SetStyle('B', true);
        if ($tag == 'A')
            $this->HREF = $attr['HREF'];
        /* if ($tag == 'BR')
          $this->Ln(5); */
        if ($tag == 'P') {
            $this->Ln(5);
            $this->SetX($posX);
        }
        if ($tag == 'PAGE') {
            $this->addPAge();
        }
    }

    function CloseTag($tag) {
        $tag = $this->convertTag($tag);
        //Closing tag
        if ($tag == 'B' or $tag == 'I' or $tag == 'U')
            $this->SetStyle($tag, false);
        if ($tag == 'H1' or $tag == 'H2' or $tag == 'H3' or $tag == 'STRONG') {
            $this->SetStyle('B', false);
        }
        if ($tag == 'A')
            $this->HREF = '';
        if ($tag == 'P' or $tag == 'H1' or $tag == 'H2' or $tag == 'H3')
            $this->Ln(5);
    }

    function SetStyle($tag, $enable) {
        //Modify style and select corresponding font
        $this->$tag += ( $enable ? 1 : -1);
        $style = '';
        foreach (array(
    'B',
    'I',
    'U'
        ) as $s)
            if ($this->$s > 0)
                $style .= $s;
        $this->SetFont('', $style);
    }

    function PutLink($URL, $txt) {
        //Put a hyperlink
        $this->SetTextColor(0, 0, 255);
        $this->SetStyle('U', true);
        $this->Write(5, $txt, $URL);
        $this->SetStyle('U', false);
        $this->SetTextColor(0);
    }

    function WriteTableCab($data, $w, $posx) {
        $this->SetX($posx);

        //	$this->SetLineWidth(.3);
        $this->SetFillColor(255, 255, 255);
        $this->SetDrawColor(192, 192, 192);
        $this->SetTextColor(0);
        $this->SetFont('');
        $fila = 0;

        $row = $data['datos'][$fila];
        $nb = 0;
        for ($i = 0; $i < count($row); $i++)
            $nb = max($nb, $this->NbLines($w[$i], trim($row[$i])));

        $h = 5 * $nb;
        //$this->Ln(1);

        for ($i = 0; $i < count($row); $i++) {
            $this->SetDrawColor(192, 192, 192);
            $x = $this->GetX();
            $y = $this->GetY();
            $align = '';
            $estilo = 'D';
            $this->SetFillColor(255, 255, 255);

            if (is_array($data['atributos'][$fila][$i]))
                $atributos = $data['atributos'][$fila][$i];
            else
                $atributos = $data['atributos'][$fila];

            $noprint = '';
            if (strpos($atributos, 'noprint') !== false) {
                $noprint = "true";
            }
            foreach ($atributos as $nattr => $attr) {
                if (strpos($attr, 'noprint') !== false) {
                    $noprint = "true";
                }

                if (strtolower($attr) == 'th') {
                    $align = 'C';
                    $estilo = 'DF';
                    $this->SetFillColor(230, 230, 230);
                    $this->SetDrawColor(0, 0, 0);
                }
            }

            if ($noprint == 'true')
                continue;
            if (is_numeric(str_replace(',', '', $row[$i])) || is_float($row[$i]))
                $align = 'R';

            $this->Rect($x, $y, $w[$i], $h, $estilo);
            $this->MultiCell($w[$i], 5, trim($row[$i]), 0, $align);
            //Put the position to the right of the cell
            $this->SetXY($x + $w[$i], $y);
        }
        $this->Ln($h);
    }

    function WriteTable($data, $w, $posx = 'X', $posy = 'X', $objCampo='') {
        $defaultFont = $this->defaultFont;
        $font = $defaultFont;
        //	$this->SetLineWidth(.3);

        if ($posx != 'X')
            $this->SetX($posx);
        else
            $this->SetX($this->offsetX);

        if ($posy != 'X')
            $this->SetY($posy);

        $this->SetFillColor(255, 255, 255);
        $this->SetDrawColor(192, 192, 192);
        $this->SetTextColor(0);
        $this->SetFont('');
        $fila = 0;
        if (isset($data['datos']))
            foreach ($data['datos'] as $nrow => $row) {

                if ($posx != 'X')
                    $this->SetX($posx);
                else
                    $posx = $this->getX();

                $nb = 0;
                $cols = count($row);

                for ($i = 0; $i < $cols; $i++)
                    $nb = max($nb, $this->NbLines($w[$i], trim($row[$i])));

                $h = ($this->fontsize / 2) * $nb;
                if ($objCampo != '') {
                    if ($objCampo->PDFpageBreak == 'false')
                        $salto = $this->CheckPageBreak($h);
                }
                else
                    $salto = $this->CheckPageBreak($h);

                if (($salto)) {
                    $this->Ln(5);
                    $this->WriteTableCab($data, $w, $posx);
                }

                for ($i = 0; $i < $cols; $i++) {
                    $this->SetDrawColor(152, 152, 152);
                    $imagen[$i] = false;

                    $x = $this->GetX();
                    $y = $this->GetY();
                    $align = 'J';
                    $estilo = 'D';
                    $estiloFont = '';
                    $borde = '0';
                    $sizeOrig = $this->fontsize;
                    $size = $sizeOrig;
                    $scalex = 100;

                    $this->SetFillColor(255, 255, 255);
                    //Recorro los Atributos
                    if (is_array($data['atributos'][$nrow][$i])) {
                        $atributos = $data['atributos'][$nrow][$i];
                    } else {
                        $atributos = $data['atributos'][$nrow];
                    }
                    $check = '-';
                    $noprint = '';

                    $atributos = $data['atributos'][$nrow][$i];
                    if (!is_array($atributos))
                        $atributos = array($atributos);

                    foreach ($atributos as $nattr => $attr) {

                        $attrib = strtolower($attr);

                        if (strpos($attrib, 'scalex=') !== false) {

                            parse_str($attrib);

                            //Start Transformation
                            $this->StartTransform();
                            $this->ScaleX($scalex);

                            $transform = true;
                        }

                        if (strpos($attrib, 'size=') !== false) {
                            parse_str($attrib);
                            $this->SetFontSize($size);
                            $size = $this->fontsize;
                            //	$nb = max($nb, $this->NbLines($w[$i], trim($row[$i])));

                            $h = ($size / 2) * $nb;
                            $sizeesp = true;
                        }
                        if (strpos($attrib, 'font=') !== false) {
                            parse_str($attr);
                            if ($this->fontsAdded[$font] != true)
                                $this->AddFont($font, '', $font . '.php');
                            $this->fontsAdded[$font] = true;
                            $this->SetFont($font);
                        }
                        if (strpos($attrib, 'noborder=') !== false) {
                            $noborder = false;
                            parse_str($attr);
                        }

                        if (strpos($attrib, 'pdffill=') !== false) {
                            $pdffill = '';
                            parse_str($attrib);
                            $this->SetFillColor($pdffill);
                        }
                        if (strpos($attrib, 'color=') !== false) {
                            $color = '';
                            $rgb = '';
                            parse_str($attrib);
                            if ($color != '')
                                $rgb = hex_to_rgb($color);
                        }

                        if (strpos($attrib, 'textcolor=') !== false) {
                            $textcolor = '';
                            $textrgb = '';
                            parse_str($attrib);
                            if ($textcolor != '')
                                $textrgb = hex_to_rgb($textcolor);
                        }


                        if (strpos($attrib, 'noprint') !== false) {
                            $noprint = "true";
                        }
                        //loger($attrib);
                        switch ($attrib) {
                            case 'anchofijochecked' :
                                $check = 'true';
                                break;
                            case 'anchofijonochecked' :
                                $check = 'false';
                                break;
                            case 'checked' :
                                $check = 'true';
                                break;
                            case 'nochecked' :
                                $check = 'false';
                                break;
                            case 'b' :
                                $estiloFont .= 'B';
                                $this->SetFont('', $estiloFont);
                                break;
                            case 'i' :
                                $estiloFont .= 'I';
                                $this->SetFont('', $estiloFont);
                                break;

                            case 'th' :
                                $align = 'C';
                                $estilo = 'DF';
                                //$this->SetDrawColor(, 0, 0);
                                $this->SetFillColor(230, 230, 230);
                                break;
                            case 'label' :
                                $align = 'L';
                                $borde = 'B';
                                $estilo = 'F';
                                $soloborde = true;
                                $this->SetDrawColor(92, 92, 92);
                                if ($pdffill == '')
                                    $this->SetFillColor(240, 240, 240);
                                break;
                            case 'bordeinf' :
                                $borde = 'B';
                                $estilo = 'F';
                                $soloborde = true;
                                break;

                            case 'sinborde' :
                                $estilo = 'F';
                                $sinborde = true;
                                $soloborde = true;
                                break;
                            case 'img' :

                                if ($atributos[$i] == 'img')
                                    $imagen[$i] = true;

                                break;
                            case 'left':
                                $align = 'L';
                                break;
                        }
                    }
                    if ($noprint == 'true')
                        continue;
                    if ($noborder == 'true') {
                        $borde = '';
                        $fill = '';
                    } else {
                        if (!($sinborde)) {
                            if ($soloborde)
                                $this->Rect($x, $y + 0.5, $w[$i], $h - 0.5, $estilo);
                            else
                                $this->Rect($x, $y, $w[$i], $h, $estilo);
                        }
                    }

                    // Si el contenido es una sola palabra y NO entra, disminuyo el tamaño de la fuente
                    //	$valor = utf8_encode(trim($row[$i]));

                    $valor = trim($row[$i]);

                    if ($objCampo != '' && $i == 1) {

                        switch ($objCampo->TipoDato) {
                            case "integer" :
                                $align = 'R';
                                break;
                            case "decimal" :
                                $align = 'R';
                                break;
                            case "numeric" :
                            case "custom_numeric" :


                                $align = 'R';
                                $precision = ($objCampo->numberPrecision != '')? $objCampo->numberPrecision:2;
                                $valor = number_format($valor, $precision, ',', '.');
                                break;
                            case "date" :
                            case "time" :
                                $align = 'C';
                                break;
                        }
                    }

                    if ($objCampo->PDFborder != '')
                        $borde = $objCampo->PDFborder;

                    if ($imagen[$i]) {
                        $valoresImagen = '';
                        parse_str($valor, $valoresImagen);

                        if ($valoresImagen['thumb_php?url'] != '') {
                            $valor = $valoresImagen['thumb_php?url'];
                            $anchoImg = $valoresImagen['ancho'] / 5;
			    if (is_file($valor)){

                                $this->Image($valor, $x + $this->offsetX, $y, $anchoImg, null, 'jpg');
			    }
                            $valor = 'Imagen';
                        }
                    }
                    $fill = 0;
                    if ($textrgb != '') {
                        $this->SetTextColor($textrgb['red'], $textrgb['green'], $textrgb['blue']);
                    }

                    $contador = 0;
                    if ($sizeesp !== true)
                        while ($this->GetStringWidth($valor) > $w[$i] && $contador < 1) {
                            $size -= 1;
                            $this->SetFont($font, $estiloFont, $size);
                            $contador++;
                        }

                    if ($rgb != '') {
                        $this->SetFillColor($rgb['red'], $rgb['green'], $rgb['blue']);
                        $this->SetDrawColor($rgb['red'], $rgb['green'], $rgb['blue']);
                        $this->Rect($x, $y, 5, 5, 'DF');
                    } else {
                        if ($check != '-') {
                            $this->SetDrawColor(0, 0, 0);
                            $this->CheckBox(($x + 1), ($y + 1), $check);
                            $this->SetDrawColor(192, 192, 192);
                        }
                        else
                            $this->MultiCell($w[$i], ($this->fontsize / 1.7), $valor, $borde, $align, $fill);
                    }

                    if ($transform)
                        $this->StopTransform();


                    $this->SetFont($defaultFont, $estiloFont, $size);
                    $this->SetFont('');
                    //Put the position to the right of the cell

                    $this->SetXY($x + $w[$i], $y);
                    $size = $sizeOrig;
                    $this->SetFontSize($size);
                }
                $this->maxY = max($this->maxY, $this->GetY());
                $this->Ln($h);

                $fila++;
            }
    }

    function NbLines($w, $txt) {
        //		$this->cMargin = 0.05;
        //Computes the number of lines a MultiCell of width w will take
        $cw = & $this->CurrentFont['cw'];
        if ($w == 0)
            $w = $this->w - $this->rMargin - $this->x;
        $wmax = ($w - 2 * $this->cMargin) * 1000 / $this->FontSize;
        $s = str_replace("\r", '', $txt);
        $nb = strlen($s);
        //		loger('txt '.$txt.' len: '.$nb, 'pdf');
        if ($nb > 0 and $s[$nb - 1] == "\n")
            $nb--;
        $sep = -1;
        $i = 0;
        $j = 0;
        $l = 0;
        $nl = 1;
        while ($i < $nb) {
            $c = $s[$i];
            if ($c == "\n") {
                $i++;
                $sep = -1;
                $j = $i;
                $l = 0;
                $nl++;
                continue;
            }
            if ($c == ' ')
                $sep = $i;
            $l += $cw[$c];
            if ($l > $wmax) {
                if ($sep == -1) {
                    if ($i == $j)
                        $i++;
                } else
                    $i = $sep + 1;
                $sep = -1;
                $j = $i;
                $l = 0;
                $nl++;
            } else
                $i++;
        }
        return $nl;
    }

    function CheckPageBreak($h, $totals = [], $fields = []) {
        //If the height h would cause an overflow, add a new page immediately
        $trigger = $this->PageBreakTrigger ;

        if (isset($this->pie)) {
            $trigger = $this->PageBreakTrigger - 12;
        }

        if ($this->footerFields != '') {
            $trigger = $this->PageBreakTrigger - abs($this->bottomMargin);
        }

        if ($this->GetY() + $h >= $trigger) {
            if (count($totals)) {
                if (implode('', $totals) != '') {
                    $this->impCabTabla($totals, $this->Columnwidths, $fields, 10, false);
                }
            }
            $this->AddPage($this->CurOrientation);
            return true;
        }
        return false;
    }

    function ReplaceHTML($html) {
        $html = str_replace('<li>', "\n<br> - ", $html);
        $html = str_replace('<LI>', "\n - ", $html);
        $html = str_replace('</ul>', "\n\n", $html);
        $html = str_replace('<strong>', "<b>", $html);
        $html = str_replace('</strong>', "</b>", $html);
        $html = str_replace('&#160;', "\n", $html);
        $html = str_replace('&nbsp;', " ", $html);
        $html = str_replace('&quot;', "\"", $html);
        $html = str_replace('&#39;', "'", $html);
        return $html;
    }

    function ParseTable($Table) {
        $_var = '';
        $_atrib = '';

        $htmlText = $Table;
        $parser = new HtmlParser($htmlText);
        while ($parser->parse()) {
            if (strtolower($parser->iNodeName) == 'table') {
                if ($parser->iNodeType == NODE_TYPE_ENDELEMENT) {
                    $_var .= '/::';
                    $_atrib .= '/::';
                } else {
                    $_var .= '::';
                    $_atrib .= '::';
                }
            }

            if (strtolower($parser->iNodeName) == 'tr') {
                if ($parser->iNodeType == NODE_TYPE_ENDELEMENT) {
                    $_var .= '!-:'; //opening row
                    $_atrib .= '!-:'; //opening row
                } else {
                    $_var .= ':-!'; //closing row
                    $_atrib .= ':-!'; //closing row
                }
            }

            if ((strtolower($parser->iNodeName) == 'td' || strtolower($parser->iNodeName) == 'th') && $parser->iNodeType == NODE_TYPE_ENDELEMENT) {
                if (strtolower($parser->iNodeName) == 'th') {
                    $_atrib .= 'TH';
                }
                $_var .= '#,#';
                $_atrib .= '#,#';
            }
            /* 				if (strpos($parser->iNodeAttributes['style'], 'display:none') !== false) {

              //				$_atrib .= 'noprint';

              } */
            if ($parser->iNodeName == 'img' && isset($parser->iNodeValue)) {
                $_var .= $parser->iNodeAttributes['src'];
                $_atrib .= 'img';
            }
            if (isset($parser->iNodeAttributes['noprint'])) {
                $_atrib .= 'noprint';
            }

            if ($parser->iNodeName == 'Text' && isset($parser->iNodeValue)) {
                $_var .= $parser->iNodeValue;
            }
            if ($parser->iNodeName == 'input' && isset($parser->iNodeValue)) {
                $_var .= $parser->iNodeAttributes['value'];
            }

            if (isset($parser->iNodeAttributes['type'])) {

                if ($parser->iNodeAttributes['type'] == 'checkbox') {
                    if ($parser->iNodeAttributes['checked'] == 'true') {
                        $_atrib .= 'checked';
                    }
                }
            }
            if (isset($parser->iNodeAttributes['pdfanchofijo'])) {

                $_atrib .= 'anchofijo';
            }
        }
        $elems = explode(':-!', str_replace('//', '', str_replace('::', '', str_replace('!-:', '', $_var)))); //opening row
        $atrib = explode(':-!', str_replace('//', '', str_replace('::', '', str_replace('!-:', '', $_atrib)))); //opening row

        foreach ($elems as $key => $value) {
            if (trim($value) != '') {
                $elems2 = explode('#,#', $value);
                array_pop($elems2);
                $data['datos'][] = $elems2;
            }
        }
        foreach ($atrib as $key => $value) {
            if (trim($value) != '') {
                $atrib2 = explode('#,#', $value);
                array_pop($atrib2);

                $data['atributos'][] = $atrib2;
            }
        }

        return $data;
    }

    function showGraficos($MisDatos) {

        if (isset($MisDatos->grafico) && $MisDatos->grafico != '')
            foreach ($MisDatos->grafico as $graf => $Objgrafico) {
                $uid = uniqid() . '.png';
                $savetemp = 'true';
                include("grafico.php");
                $x = $this->GetX();
                $y = max($this->maxY, $this->GetY());
                //			echo $y;
                $dataPath = $_SESSION['datapath'];

                if ($dataPath != '') {
                    $tmpbase = '../database/' . $dataPath;
                }
		if (is_file($tmpbase . '/tmp/' . $uid)){

            	    $this->Image($tmpbase . '/tmp/' . $uid, $x + $this->offsetX, $y, 100, null, 'png');
		}
                unlink($tmpbase . '/tmp/' . $uid);
            }
    }

    function WriteHTML($html, $posx = 'X') {

        $html = $this->ReplaceHTML($html);
        //Search for a table
        $start = strpos(strtolower($html), '<table');
        $end = strrpos(strtolower($html), '</table');

        if ($start !== false && $end !== false) {
            $this->WriteHTML2(substr($html, 0, $start));

            $tableVar = substr($html, $start, $end - $start);
            $tableData = $this->ParseTable($tableVar);

            $min = 0;
            $max = 0;
            $w = '';
            $columnas = 0;
            foreach ($tableData['datos'] as $nfila => $row) {

                $anchofila = 0;
                foreach ($row as $ncol => $valor) {
                    if (strstr($tableData['atributos'][$nfila][$ncol], 'noprint') !== false)
                        continue;

                    $columnas++;
                    if (strstr($tableData['atributos'][$nfila][$ncol], 'C=') != '') {
                        //	echo $tableData['atributos'][$nfila][$ncol];
                    } else
                        $w[$ncol] = max($this->GetStringWidth($valor), $w[$ncol]);
                    if (strstr($tableData['atributos'][$nfila][$ncol], 'anchofijo') != '') {
                        //$fija[$ncol] = max($w[$ncol], $fija[$ncol]);
                    }
                }
            }
            // Obtengo las Columnas FIJAS y a esas NO LAS COMPENSO
            // suavizo las curvas de valores */
            // con una constante de compensacion

            $compensacion = (count($w) - count($fija)) / 4;
            $compensacion = (count($w)) / 4;
            $min = min($w);
            $max = max($w);

            foreach ($w as $n => $valor) {
                // No compenso las fijas
                if ($fija[$n] != '') {
                    //las fijas
                } else {
                    if ($max - $valor >= $valor - $min)
                        $w[$n] = $valor + $compensacion;
                    else
                        $w[$n] = $valor - $compensacion;
                }
            }

            // ancho total
            $anchofilafin = array_sum($w);
            $anchopag = $this->anchoPagina;
            if ($posx != 'X')
                $anchopag = $this->anchoPagina - $posx;

            /* calculo proporciones */
            for ($i = 0; $i < $columnas; $i++) {
                if ($fija[$i] != '') {
                    // las fijas
                } else {
                    $porcelda = ($w[$i] / $anchofilafin) * 100;
                    $w[$i] = ($anchopag / 100) * $porcelda;
                }
            }
            $this->WriteTable($tableData, $w, $posx);

            $this->WriteHTML2(substr($html, $end + 8, strlen($html) - 1) . '<BR>');
        } else {
            $this->WriteHTML2($html);
        }
        $this->maxY = max($this->maxY, $this->GetY());
    }

    function wordcount($str) {
        return count(explode(" ", $str));
    }

    //Cabecera de página
    function Header() {
//	$this->pageNumber++;
        $offX = $this->margenX / 2;
        
        if ($this->customHeader()) {
	    $this->sincab = true;
        }

        // CUSTOM HEADER
        if ($this->headerFields != ''){
            
            $this->SetFont($this->defaultFont, 'BI', 7);
            foreach($this->headerFields as $field){
                $this->labelXY($field, null , $this->Container );
            }
        }
        
        
        
        if ($this->sincab)
            return;

        $datosbase = $this->datosbase;

        $datapath = '../database/' . $datosbase->xmlPath;

        /* Tomo los datos de cada conexion */
        $img_fondo = $datosbase->img_fondo;
        $nom_empresa = $datosbase->nombre;
        $direccion = $datosbase->direccion;
        $cuit = $datosbase->cuit;
        $telefonos = $datosbase->telefonos;
        $logo_pdf_1 = $datosbase->logo_pdf_1;
        $logo_pdf_1_attr = $datosbase->logo_pdf_1_attr;
        
        $logo_pdf_2 = $datosbase->logo_pdf_2;
        $logo_pdf_2_attr = $datosbase->logo_pdf_2_attr;

        // Empresa
        $empnom = utf8_decode($nom_empresa);
        $empdir = utf8_decode($direccion);
        if ($cuit != '')
            $empcuit = utf8_decode('Cuit: ' . $cuit);
        $emptel = utf8_decode($telefonos);


        //Logo

        if (is_file($datapath . 'img/' . $logo_pdf_1)){
        	$posx = ($logo_pdf_1_attr['posx'] != '')?$logo_pdf_1_attr['posx']:14;
        	$posy = ($logo_pdf_1_attr['posy'] != '')?$logo_pdf_1_attr['posy']:3;
        	$width = ($logo_pdf_1_attr['width'] != '')?$logo_pdf_1_attr['width']:20;        	
            $this->Image($datapath . 'img/' . $logo_pdf_1, $posx + $this->offsetX, $posy + $this->offsetY, $width);
        }
        
        if (is_file($datapath . 'img/' . $logo_pdf_2)){
            $posx = ($logo_pdf_2_attr['posx'] != '')?$logo_pdf_2_attr['posx']:38;
        	$posy = ($logo_pdf_2_attr['posy'] != '')?$logo_pdf_2_attr['posy']:3;        	   
        	$width = ($logo_pdf_2_attr['width'] != '')?$logo_pdf_2_attr['width']:20;   
            $this->Image($datapath . 'img/' . $logo_pdf_2, $posx + $this->offsetX, $posy + $this->offsetY, $width);
        }


        // EMPRESA
        $this->SetFont($this->defaultFont, 'BI', 7);
        $lineY = 11 + $this->offsetY;
        $this->setXY(4 + $this->offsetX, $lineY);

        $width = $this->GetStringWidth($empnom) * 1.4;
        $this->MultiCell($width, 4, $empnom, 0, 'L');


        //DIRECCION y Otros
        $this->SetFont($this->defaultFont, 'I', 7);
        $this->setXY($width + $this->offsetX, $lineY + 1.5);

        $strOtros = $empdir . ' ' . $empcuit . ' ' . $emptel;

        $width = $this->anchoPagina - $this->GetX();

        $h = $this->NbLines($width, $strOtros);

        $this->MultiCell($width, $h, $strOtros, 0, 'L');

        $maxY = $this->GetY();

        $this->SetFont($this->defaultFont, 'BI', 10);

        // Fecha
        $nodate = (isset($this->sinFecha)) ? $this->sinFecha : 'false';

        $this->SetFont($this->defaultFont, '', 9);
        $fecha = utf8_decode( $this->i18n['printDate'].':' . date('d/m/Y'));
        $anchoFecha = $this->GetStringWidth($fecha) + 3;
        $anchorecuadro = $anchoFecha + 2;
        $posRecuadro = $this->offsetX + $this->anchoPagina - $anchorecuadro;

        if ($nodate != 'true') {

            $this->RoundedRect($posRecuadro, 3 + $this->offsetY, $anchoFecha, 8, 2);
            $this->setXY($posRecuadro + 1, 5 + $this->offsetY);
            $this->Cell(0, 0, $fecha, 0, 0, 'L');
        } 

        $nonum = (isset($this->sinNumero)) ? $this->sinNumero : 'false';
        $hoja = '';
        if ($nonum != 'true') {
            $hoja = $this->i18n['page'] .': '. $this->PageNo() ;
            if ($this->pageNumber == 0){
                $hoja .= '/{nb}';
            }
            $this->setXY($posRecuadro + 1, 8 + $this->offsetY);
            $this->Cell(0, 0, $hoja, 0, 0, 'L');
        }

        // USER
        $this->SetFont($this->defaultFont, '', 3);
        $this->setXY($this->anchoPagina - 1, $this->offsetY + 1);
        $this->Cell(0, 0, $this->user, 0, 0, 'L');

        $anchoHoja = $this->GetStringWidth($hoja) + 3;

        //Arial bold 15
        $this->SetFont($this->defaultFont, 'B', 15);
        //Movernos a la derecha
        //$this->Cell(50);
        //Título
        $maxHeight = $maxY - $this->offsetY - 0.5;
        $this->RoundedRect($offX + $this->offsetX, 2 + $this->offsetY, $this->anchoPagina - 2, $maxHeight, 2);
        $this->setXY($this->offsetX, 2 + $this->offsetY);
        $this->titulo = strip_tags($this->titulo);
        $this->Cell(0, 10, $this->titulo, 0, 0, 'C');
        $this->Ln(1);
        //	Salto de línea
        //$this->setXY(0 + $this->offsetX, 22 + $this->offsetY);

        $this->maxY = $maxHeight + $this->offsetY + 3;
        $this->setXY($this->offsetX, $this->maxY);

        $this->maxHeaderY = $this->maxY;
    }

    function customHeader()
    {
        $pdf = $this;
        $code = isset($this->Container->pdfHeader)?$this->Container->pdfHeader:'';
        if ($code != '') {
            eval($code);
            $this->maxY = $this->GetY();
            $this->maxHeaderY = $this->maxY;
            return true;
        }
        return false;
    }
    
    function customFooter()
    {
        $pdf = $this;
        $code = isset($this->Container->pdfFooter)?$this->Container->pdfFooter:'';
        if ($code != '') {
            eval($code);
            return true;
        }
        return false;
    }

    function pageNo(){
        $offset = $this->pageNumber;
        $page = parent::PageNo();
        return $page + $offset ;
    }
    //Pie de página
    function Footer() {

        if ($this->customFooter()) {
	    $this->sincab = true;
        }

        // CUSTOM Footer fields
        if ($this->footerFields != ''){
            $this->SetFont($this->defaultFont, 'BI', 7);
            foreach($this->footerFields as $field){
		$field->checkPageBreak = 'false';
                $this->labelXY($field, null , $this->Container );
            }
        }

        if (!$this->sincab) {  // && !isset($this->pie)  

            $supportName = isset($_SESSION['properties']['supportName']) ? $_SESSION['properties']['supportName'] : 'www.histrix.com.ar';
            $supportUrl  = isset($_SESSION['properties']['supportUrl' ]) ? $_SESSION['properties']['supportUrl']  : 'http://www.estudiogenus.com/histrix';

            $this->SetY(-5);
            $this->SetTextColor(192, 192, 192);            
            $this->SetFont($this->defaultFont, 'B', 6);            
            $this->Cell(0, 0, $supportName, 0, 0, 'L', null, $supportUrl);

            $this->SetFont($this->defaultFont, 'B', 15);
            $this->SetTextColor(0);            
        }

        if ($this->sincab && $this->conpie != 'true')
            return;

        //Posición: a 1,5 cm del final
        $this->bottomMargin = -12;
        //Arial italic 8
        $this->SetFont($this->defaultFont, 'I', 8);
        //Número de página

        if (isset($this->pie)) {
            $this->SetY($this->bottomMargin);

            $size = ($this->pie['size'] != '') ? $this->pie['size'] : 9;
            $style = ($this->pie['style'] != '') ? $this->pie['style'] : 'I';
            $align = ($this->pie['align'] != '') ? $this->pie['align'] : 'L';

            if ($this->pie['color'] != '') {
                $textrgb = hex_to_rgb($this->pie['color']);
                $this->SetTextColor($textrgb['red'], $textrgb['green'], $textrgb['blue']);
            }


            $this->SetFont($this->defaultFont, $style, $size);

            $this->SetX(6);
            $pietext = utf8_decode($this->pie['text']);

            $w = $this->GetStringWidth($pietext);
            $h = $this->NbLines($w, $pietext) * 2;

            //$this->SetY(-8 * ($h / 2)) ;

            $this->MultiCell(0, $h, $pietext, 0, $align);

            $this->SetTextColor(0, 0, 0);
        } else {
            //	if ($this->sinNumero != 'true') {
            //		$this->Cell(0, 10, 'Hoja ' . $this->PageNo() . '/{nb}', 0, 0, 'C');
            //	}
        }
    }

    /*
      x, y: top left corner of the rectangle.
      w, h: width and height.
      r: radius of the rounded corners.
      style: same as Rect(): F, D (default), FD or DF.

     */

    function RoundedRect($x, $y, $w, $h, $r, $style = '') {
        $k = $this->k;
        $hp = $this->h;
        if ($style == 'F')
            $op = 'f';
        elseif ($style == 'FD' or $style == 'DF')
            $op = 'B';
        else
            $op = 'S';
        $MyArc = 4 / 3 * (sqrt(2) - 1);
        $this->_out(sprintf('%.2f %.2f m', ($x + $r) * $k, ($hp - $y) * $k));
        $xc = $x + $w - $r;
        $yc = $y + $r;
        $this->_out(sprintf('%.2f %.2f l', $xc * $k, ($hp - $y) * $k));

        $this->_Arc($xc + $r * $MyArc, $yc - $r, $xc + $r, $yc - $r * $MyArc, $xc + $r, $yc);
        $xc = $x + $w - $r;
        $yc = $y + $h - $r;
        $this->_out(sprintf('%.2f %.2f l', ($x + $w) * $k, ($hp - $yc) * $k));
        $this->_Arc($xc + $r, $yc + $r * $MyArc, $xc + $r * $MyArc, $yc + $r, $xc, $yc + $r);
        $xc = $x + $r;
        $yc = $y + $h - $r;
        $this->_out(sprintf('%.2f %.2f l', $xc * $k, ($hp - ($y + $h)) * $k));
        $this->_Arc($xc - $r * $MyArc, $yc + $r, $xc - $r, $yc + $r * $MyArc, $xc - $r, $yc);
        $xc = $x + $r;
        $yc = $y + $r;
        $this->_out(sprintf('%.2f %.2f l', ($x) * $k, ($hp - $yc) * $k));
        $this->_Arc($xc - $r, $yc - $r * $MyArc, $xc - $r * $MyArc, $yc - $r, $xc, $yc - $r);
        $this->_out($op);
    }

    function _Arc($x1, $y1, $x2, $y2, $x3, $y3) {
        $h = $this->h;
        $this->_out(sprintf('%.2f %.2f %.2f %.2f %.2f %.2f c ', $x1 * $this->k, ($h - $y1) * $this->k, $x2 * $this->k, ($h - $y2) * $this->k, $x3 * $this->k, ($h - $y3) * $this->k));
    }

    function CheckBox($x, $y, $check, $ancho=3) {
        $this->Rect($x, $y, $ancho, $ancho);
        if ($check == 'true' || $check == 1) {
            $medioancho = $ancho / 2;
            $cuartoancho = $ancho / 4;
            $xcheck1 = $x + $cuartoancho;
            $ycheck1 = $y + $medioancho;
            $xcheck2 = $x + $medioancho;
            $ycheck2 = $y + $ancho * 3 / 4;
            $xcheck3 = $x + ($ancho * 4 / 5);
            $ycheck3 = $y + $cuartoancho;
            $this->Line($xcheck1, $ycheck1, $xcheck2, $ycheck2);
            $this->Line($xcheck2, $ycheck2, $xcheck3, $ycheck3);
        }
    }

    function DashedRect($x1, $y1, $x2, $y2, $width = 1, $nb = 15) {
        $this->SetLineWidth($width);
        $longueur = abs($x1 - $x2);
        $hauteur = abs($y1 - $y2);
        if ($longueur > $hauteur) {
            $Pointilles = ($longueur / $nb) / 2; // length of dashes
        } else {
            $Pointilles = ($hauteur / $nb) / 2;
        }
        for ($i = $x1; $i <= $x2; $i += $Pointilles + $Pointilles) {
            for ($j = $i; $j <= ($i + $Pointilles); $j++) {
                if ($j <= ($x2 - 1)) {
                    $this->Line($j, $y1, $j + 1, $y1); // upper dashes
                    $this->Line($j, $y2, $j + 1, $y2); // lower dashes
                }
            }
        }
        for ($i = $y1; $i <= $y2; $i += $Pointilles + $Pointilles) {
            for ($j = $i; $j <= ($i + $Pointilles); $j++) {
                if ($j <= ($y2 - 1)) {
                    $this->Line($x1, $j, $x1, $j + 1); // left dashes
                    $this->Line($x2, $j, $x2, $j + 1); // right dashes
                }
            }
        }
    }

    /**
     * HACER QUE SOLO IMPRIMA LAS RAMAS ABIERTAS DE LOS ARBOLES
     * variable de SESSION['ARBOL'];
     */
//	var $numNod;

    function showTree($MisDatos, $tree, $ident, $margen, $root=false, $nivel=0) {
        $activos = explode('.', $_SESSION['ARBOL']);

        //$activos[]=0;
        $paso = 3;
        $ident += $paso * 2;
        $treeRoot = $tree->nodos;


        $currentY2 = $this->GetY() - ($paso / 1.2);
        $show = false;

        $imagen = '../img/folderopen.jpg';
        if (is_array($treeRoot)) {


            foreach ($treeRoot as $order => $node) {
                $opt['titulo'] = false;
                $opt['tree'] = true;
                $lineHX = $ident - $paso + $margen;
                //$lineHX= $ident - ($paso * 2.5) + $margen;

                if (in_array($this->numNod , $activos))
                    $show = true;
                
                if ($root)
                    $show = true;

                $this->numNod++;

                if ($show == true) {
                    $this->ln(0.5);

                    $maxY = max($currentY, $maxY);
                    $minY = min($currentY, $minY);
                    $Height = 0;

                    $this->CheckPageBreak($paso * 2);
                    $currentY = $this->GetY() + ($paso / 2);


                    $this->Line($lineHX, $currentY, $ident + $margen, $currentY); // Horizontal Line

                    $Height = $this->impTabla($MisDatos, array($node->dataRow), $opt, 'auto', null, $ident + $margen);
                    $this->setXY($lineHX, $currentY + $Height);

                    $this->Line($lineHX, $currentY, $lineHX, $currentY - $Height - $next); // Vartical Line
                    $next = $Height / 2;
                    $niv = $nivel;
                    $posx = $lineHX;
                    while ($niv) {
                        $posx = $posx - ( $paso * 2);
                        $this->Line($posx, $currentY + $Height / 2, $posx, $currentY - $Height * 1.5); // Vertical
                        $niv--;
                    }

                    //$this->Image($imagen, $lineHX - 1 , $currentY - 1 , 3 , 3);
                    //$this->Cell(10,2,$this->numNod);
                }

                $maxY2 = $this->showTree($MisDatos, $node, $ident, $margen, false, $nivel + 1);
                
            }
        }
        //	if ($show)
        //   	$this->Line($lineHX, $maxY, 	$lineHX			,  $currentY2	 );
        return $maxY;
    }

    //Cell with horizontal scaling if text is too wide
    function CellFit($w, $h = 0, $txt = '', $border = 0, $ln = 0, $align = '', $fill = 0, $link = '', $scale = 0, $force = 1) {
        //Get string width
        $str_width = $this->GetStringWidth($txt);

        //Calculate ratio to fit cell
        if ($w == 0)
            $w = $this->w - $this->rMargin - $this->x;
        $ratio = ($w - $this->cMargin * 2) / $str_width;

        $fit = ($ratio < 1 || ($ratio > 1 && $force == 1));
        if ($fit) {
            switch ($scale) {

                //Character spacing
                case 0 :
                    //Calculate character spacing in points
                    $char_space = ($w - $this->cMargin * 2 - $str_width) / max($this->MBGetStringLength($txt) - 1, 1) * $this->k;
                    //Set character spacing
                    $this->_out(sprintf('BT %.2f Tc ET', $char_space));
                    break;

                //Horizontal scaling
                case 1 :
                    //Calculate horizontal scaling
                    $horiz_scale = $ratio * 100.0;
                    //Set horizontal scaling
                    $this->_out(sprintf('BT %.2f Tz ET', $horiz_scale));
                    break;
            }
            //Override user alignment (since text will fill up cell)
            $align = '';
        }

        //Pass on to Cell method
        $this->Cell($w, $h, $txt, $border, $ln, $align, $fill, $link);

        //Reset character spacing/horizontal scaling
        if ($fit)
            $this->_out('BT ' . ($scale == 0 ? '0 Tc' : '100 Tz') . ' ET');
    }

    //Cell with horizontal scaling only if necessary
    function CellFitScale($w, $h = 0, $txt = '', $border = 0, $ln = 0, $align = '', $fill = 0, $link = '') {
        $this->CellFit($w, $h, $txt, $border, $ln, $align, $fill, $link, 1, 0);
    }

    //Cell with horizontal scaling always
    function CellFitScaleForce($w, $h = 0, $txt = '', $border = 0, $ln = 0, $align = '', $fill = 0, $link = '') {
        $this->CellFit($w, $h, $txt, $border, $ln, $align, $fill, $link, 1, 1);
    }

    //Cell with character spacing only if necessary
    function CellFitSpace($w, $h = 0, $txt = '', $border = 0, $ln = 0, $align = '', $fill = 0, $link = '') {
        $this->CellFit($w, $h, $txt, $border, $ln, $align, $fill, $link, 0, 0);
    }

    //Cell with character spacing always
    function CellFitSpaceForce($w, $h = 0, $txt = '', $border = 0, $ln = 0, $align = '', $fill = 0, $link = '') {
        //Same as calling CellFit directly
        $this->CellFit($w, $h, $txt, $border, $ln, $align, $fill, $link, 0, 1);
    }

    //Patch to also work with CJK double-byte text
    function MBGetStringLength($s) {
        if ($this->CurrentFont['type'] == 'Type0') {
            $len = 0;
            $nbbytes = strlen($s);
            for ($i = 0; $i < $nbbytes; $i++) {
                if (ord($s[$i]) < 128)
                    $len++;
                else {
                    $len++;
                    $i++;
                }
            }
            return $len;
        } else
            return strlen($s);
    }

    /*     * ********** TRANSFOMATION CLASS */

    function StartTransform() {
        //save the current graphic state
        $this->_out('q');
    }

    function ScaleX($s_x, $x='', $y='') {
        $this->Scale($s_x, 100, $x, $y);
    }

    function ScaleY($s_y, $x='', $y='') {
        $this->Scale(100, $s_y, $x, $y);
    }

    function ScaleXY($s, $x='', $y='') {
        $this->Scale($s, $s, $x, $y);
    }

    function Scale($s_x, $s_y, $x='', $y='') {
        if ($x === '')
            $x = $this->x;
        if ($y === '')
            $y = $this->y;
        if ($s_x == 0 || $s_y == 0)
            $this->Error('Please use values unequal to zero for Scaling');
        $y = ($this->h - $y) * $this->k;
        $x*=$this->k;
        //calculate elements of transformation matrix
        $s_x/=100;
        $s_y/=100;
        $tm[0] = $s_x;
        $tm[1] = 0;
        $tm[2] = 0;
        $tm[3] = $s_y;
        $tm[4] = $x * (1 - $s_x);
        $tm[5] = $y * (1 - $s_y);
        //scale the coordinate system
        $this->Transform($tm);
    }

    function MirrorH($x='') {
        $this->Scale(-100, 100, $x);
    }

    function MirrorV($y='') {
        $this->Scale(100, -100, '', $y);
    }

    function MirrorP($x='', $y='') {
        $this->Scale(-100, -100, $x, $y);
    }

    function MirrorL($angle=0, $x='', $y='') {
        $this->Scale(-100, 100, $x, $y);
        $this->Rotate(-2 * ($angle - 90), $x, $y);
    }

    function TranslateX($t_x) {
        $this->Translate($t_x, 0, $x, $y);
    }

    function TranslateY($t_y) {
        $this->Translate(0, $t_y, $x, $y);
    }

    function Translate($t_x, $t_y) {
        //calculate elements of transformation matrix
        $tm[0] = 1;
        $tm[1] = 0;
        $tm[2] = 0;
        $tm[3] = 1;
        $tm[4] = $t_x * $this->k;
        $tm[5] = -$t_y * $this->k;
        //translate the coordinate system
        $this->Transform($tm);
    }

    function Rotate($angle, $x='', $y='') {
        if ($x === '')
            $x = $this->x;
        if ($y === '')
            $y = $this->y;
        $y = ($this->h - $y) * $this->k;
        $x*=$this->k;
        //calculate elements of transformation matrix
        $tm[0] = cos(deg2rad($angle));
        $tm[1] = sin(deg2rad($angle));
        $tm[2] = -$tm[1];
        $tm[3] = $tm[0];
        $tm[4] = $x + $tm[1] * $y - $tm[0] * $x;
        $tm[5] = $y - $tm[0] * $y - $tm[1] * $x;
        //rotate the coordinate system around ($x,$y)
        $this->Transform($tm);
    }

    function SkewX($angle_x, $x='', $y='') {
        $this->Skew($angle_x, 0, $x, $y);
    }

    function SkewY($angle_y, $x='', $y='') {
        $this->Skew(0, $angle_y, $x, $y);
    }

    function Skew($angle_x, $angle_y, $x='', $y='') {
        if ($x === '')
            $x = $this->x;
        if ($y === '')
            $y = $this->y;
        if ($angle_x <= -90 || $angle_x >= 90 || $angle_y <= -90 || $angle_y >= 90)
            $this->Error('Please use values between -90° and 90° for skewing');
        $x*=$this->k;
        $y = ($this->h - $y) * $this->k;
        //calculate elements of transformation matrix
        $tm[0] = 1;
        $tm[1] = tan(deg2rad($angle_y));
        $tm[2] = tan(deg2rad($angle_x));
        $tm[3] = 1;
        $tm[4] = -$tm[2] * $y;
        $tm[5] = -$tm[1] * $x;
        //skew the coordinate system
        $this->Transform($tm);
    }

    function Transform($tm) {
        $this->_out(sprintf('%.3f %.3f %.3f %.3f %.3f %.3f cm', $tm[0], $tm[1], $tm[2], $tm[3], $tm[4], $tm[5]));
    }

    function StopTransform() {
        //restore previous graphic state
        $this->_out('Q');
    }

    /**
     * Pdf Loger Method
     * @param string $message
     */                              
    function loger($message, $pX=1, $pY=1) {
        $x = $this->getX();
        $y = $this->getY();
        $this->setXY($pX, $pY);
        $this->Write( 5, print_r($message, true));
        $this->setXY($x, $y);
    }

}

?>
