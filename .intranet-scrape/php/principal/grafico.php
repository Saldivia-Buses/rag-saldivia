<?php
/**
 * Wrap para graficos
 */
$DirectAccess = true;
include_once("./sessionCheck.php");
include_once("../lib/pChart/pChart/pData.class");
include_once("../lib/pChart/pChart/pChart.class");

if ($graf =='')
    $graf = $_GET['grafico'];



$grafico = $_SESSION[$graf];
$savetmp = $_GET['savetmp'];
$ancho   = $grafico['ancho'];
$alto    = $grafico['alto'];



//print_r($grafico);
if ($_GET['ancho'] != '') 
    $ancho =$_GET['ancho'];

if (isset($grafico['objective'])) 
    $objective =$grafico['objective'];


if ($_GET['alto'] != '') 
    $alto =$_GET['alto'];

if ( count($grafico['series']) > 0 || count($grafico['valores']) > 0) {

if ($ancho == '')  $ancho = 320;

if ($alto == '')
    $alto = $ancho * 2 / 3;
    
//$alto = $alto * 2;
/*
    if ( $_GET['action'] == "GetImageMap" ){
         $mygraph = new pChart($ancho,$alto);
         $mygraph->getImageMap('IMG'.$graf);
    }
*/

    $margenV = 2;
    $margenH = 25;

    $DataSet = new pData;
    $mygraph = new pChart($ancho,$alto);
//    $mygraph->drawBackground(255,255,3);

    $mygraph->ErrorFontName ="../lib/pChart/Fonts/tahoma.ttf";
    
    //$mygraph->reportWarnings("GD");

    $fontSize = $alto / 25;


    $mygraph->drawFilledRoundedRectangle(0, $margenV,
        $ancho ,  $alto,
        5, 240, 240, 240);

    $mygraph->drawRoundedRectangle( 2 , $margenV + 2 ,
        $ancho - 2 , $alto -  2,5,230,230,230);

    $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf", $fontSize);
    // $mygraph->setShadowProperties(1,1,0,0,0);
    if ($grafico['titulo'] != ''){
        $margenV += 15;
        $mygraph->drawTitle(0, 2,$grafico['titulo'],0,0,0, $ancho, $margenV);
        
    }
    if ($grafico['subtitulo'] != ''){
        
        $mygraph->drawTitle(0, $margenV + 4 ,$grafico['subtitulo'],100,100,100, $ancho, $margenV + 25);
        $margenV += 15;
    }
    $DataSet->SetAbsciseLabelSerie();

    $i = 0;
    $cantValores = 0;
    $min = 0;
    $max=0;
    $skip = true;
    // Para cuando solo tengo 1 serie y es del tipo 'Pie'

    if ($grafico['tipo'] == 'P' ) {
        if (count($grafico['series']) > 1 ) {
          /*  echo '<pre>';
        print_r($grafico);
        echo '</pre>';
        */
            unset($valores);
            unset($titulo);
            foreach($grafico['series'] as $serie) {
                if (is_array($serie)){
		    $valSerie = array_sum($serie);
                    $valores[]= $valSerie;
		}
		if (is_array($valores))
        		$totalPie= array_sum($valores);
		    
                $titulo[]= $serie['titulo'].'('.$valSerie.')';

            }
            if (is_array($valores) != '') {
	    
                $skip = false;
                $DataSet->AddPoint( $valores ,"Serie1");
                $DataSet->AddPoint( $titulo ,"_titulo");
            }
        } else {
            $skip = false;
	    if ($grafico['valores'] != null){
	    
                $DataSet->AddPoint( $grafico['valores'] ,"Serie1");
//		$TotalvalSerie = array_sum($grafico['valores']);

	    }

	    if ($grafico['leyendas'] != ''){
        	foreach($grafico['leyendas'] as $n => $ley) {
            	    $leyendas[]=$ley.' ('.$grafico['valores'][$n].')';
    		    if ($n > 7)
    			$mygraph->setColorPalette($n - 1, rand(0,200) , rand(0,200) ,rand(0,200));
		
                }
                if (is_array($grafico['valores']))
		$totalPie= array_sum($grafico['valores']);
	    }
	    
        }

          
        //
	if ($leyendas != '')
            $DataSet->AddPoint($leyendas ,'_titulo');
        $DataSet->SetAbsciseLabelSerie("_titulo");

    } else {
//    var_dump($grafico);
    // just series but Pie charts
        if ($grafico['series'] !='' ) {

            if ($grafico['leyendas'] != ''){
                foreach($grafico['leyendas'] as $n => $ley) {
                    $leyendas[]=$ley;
                    // tags
                    //$Test->setFontProperties("Fonts/tahoma.ttf",8);

                }

                if ($leyendas != '');
                $DataSet->AddPoint($leyendas ,'_titulo');
                $DataSet->SetAbsciseLabelSerie("_titulo");
            }

            foreach($grafico['series'] as $nomcolseries => $valserie) {
                
                if ($nomcolseries != ''){
                    $seriesNames[] = $nomcolseries;
                    $DataSet->SetSerieName($valserie['titulo'], $nomcolseries);
                }

                //		#  $DataSet->SetYAxisName("Average age");
                $i++;
		$mygraph->setColorPalette($i - 1, rand(0,200) , rand(0,200) ,rand(0,200));

                if (is_array($valserie)) {
                    $c =0;
                    foreach($valserie as $ind => $vals) {
                        if ($ind === 'titulo') continue;
                        
                        $c++;
                        $cantValores++;
                        $min = min($min, $vals);
                        $max = max($max, $vals);
                        
                        $valores[$ind ]=$vals;
                    }
                    
                    if (count($grafico['series']) == 1){
                        $media = round(array_sum($valserie) / $c, 3);
			if ($media > 1 )
                            $media = round($media,0);


//                        $media = ($max - $min) / 2;
                        }
                    $skip = false;
            	    $DataSet->AddPoint($valores,$nomcolseries);
                    $DataSet->AddSerie($nomcolseries);
                /*    $Data = $DataSet->GetData();
                    $descriptions = $DataSet->GetDataDescription();

                    $mygraph->setLabel($Data,$descriptions,$nomcolseries,"Febrero",$valores[1],221,230,174);
*/

                }


            }

        }

     }

    //	    $DataSet->SetAbsciseLabelSerie("Serie2");
    // Get Data
    $Data = $DataSet->GetData();
    $descriptions = $DataSet->GetDataDescription();

    
   // if (is_array($Data[0])) $skip = true;
    //print_r($descriptions);
    $graphType = $grafico['tipo'];
    if ($graphType != 'P' ) {


        $DataSet->SetAbsciseLabelSerie('_titulo');
        $mygraph->setGraphArea( $ancho / 10 ,  $alto / 9,
            $ancho - ($ancho / 20) , $alto - ($alto / 10) );
      //  $mygraph->drawGraphArea(255,255,255,TRUE);
        $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf",$alto / 35);
        
    
	if (isset($objective) && $max != 0 && $max < $objective){
	    $mygraph->setFixedScale(0,$objective + 10); 
	} 
	//else {
    	    if ($max != $min && $max != 0)
        	$mygraph->drawScale($Data,$descriptions,
            	     SCALE_START0,150,150,150,TRUE,0,2,TRUE);
        //}   
                
        $mygraph->drawGrid(4,TRUE,230,230,230);

        $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf",$alto / 35);
        if ($grafico['leyendas'] != ''){
            $mygraph->drawLegend($margenH * 3.5 , ($margenV * 1  ) ,  $descriptions,250,250,250, 255,255,255, 0, 0, 0, true);
        }

    }
    if (is_array($Data) &&  (!$skip))
        switch($grafico['tipo']) {
            case 'P':
                $x1 = $margenH / 2;
                $y1 = ($alto / $margenV) + $margenV * 2;
                $x2 = $ancho - $margenH - $margenH ;
                $y2 = $alto - $margenV * 2;

                $DataSet->AddAllSeries();
                $Data = $DataSet->GetData();
                $descriptions = $DataSet->GetDataDescription();

                $mygraph->setGraphArea( $x1 , $y1 ,$x2, $y2);
                //		$mygraph->drawGraphArea(255,255,255,TRUE);
                $centroX = (($x2 - $x1) / 2 ) +   $margenH / 3;
                $centroX = $ancho / 3;
		
		
                $centroY = (($y2 - $y1) / 2 ) +   $margenV * 2;

$mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf",$alto / 35);
                $mygraph->drawPieGraph($Data, $descriptions, $centroX,
                    $centroY ,$ancho / 4, PIE_PERCENTAGE, true );

	        $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf",$alto / 30);
                    
                $mygraph->drawPieLegend($ancho / 2 + ($ancho / 6),($margenV),
                    $Data, $descriptions ,250,250,250);
	        $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf",$alto / 20);
		    
		$mygraph->drawTextBox($MARGENh ,$alto - $margenV * 2 , $ancho - $margenH * 2 ,$alto, 'Total: '.$totalPie,0,100 ,100,100,ALIGN_CENTER,TRUE,250,250,250,30);  

                break;
            case 'BC':
                $mygraph->drawCubicCurve($Data, $descriptions);
            case 'SC':
                $mygraph->drawStackedBarGraph($Data,$descriptions);
            case 'OC':
                $mygraph->drawOverlayBarGraph($Data,$descriptions);
   
            case 'C':
                $mygraph->drawBarGraph($Data, $descriptions,TRUE);
    		if(isset($media))
			$mygraph->drawTreshold($media,255,0,0, true, false, 4,     'Promedio '.$media); 

    		if(isset($objective))
			$mygraph->drawTreshold($objective,0,180,0, true, false, 4, 'Objetivo '.$objective); 

                break;
            case 'L':

                $mygraph->drawFilledLineGraph($Data, $descriptions, 30);

                $mygraph->drawPlotGraph($Data, $descriptions,
                    3,2,255,255,255);

                $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf", $alto / 40);
                $mygraph->writeValues($Data,'     '.$descriptions,$seriesNames);

        		if(isset($media))
    			     $mygraph->drawTreshold($media,255,0,0, true, false, 4,     'Promedio '.$media); 

        		if(isset($objective))
    		  	   $mygraph->drawTreshold($objective,0,180,0, true, false, 4, 'Objetivo  '.$objective); 

                //$mygraph->drawFilledLineGraph($DataSet->GetData(),
                //                  $DataSet->GetDataDescription(),50,TRUE);

                break;
            case 'UL':

                $mygraph->drawLineGraph($Data, $descriptions);
          /*      $mygraph->drawPlotGraph($Data, $descriptions,
                    3,2,255,255,255);
*/
                $mygraph->setFontProperties("../lib/pChart/Fonts/tahoma.ttf", 7);
                $mygraph->writeValues($Data,$descriptions,$seriesNames);

                //$mygraph->drawFilledLineGraph($DataSet->GetData(),
                //                  $DataSet->GetDataDescription(),50,TRUE);
                break;

        }
    //$mygraph->drawBarGraph($DataSet->GetData(),
    //                      $DataSet->GetDataDescription(),TRUE);

    $dataPath = $_SESSION['datapath'];

    if ($dataPath != '') {
        $tmpbase= '../database/'.$dataPath;
    }
                                                         
    if ($savetemp ==  'true') {

        $mygraph->Render( $tmpbase.'/tmp/'.$uid);
    }
    else {

        $mygraph->Render( $tmpbase.'/tmp/'.$graf.'.png');

        $mygraph->Stroke();

    }
}
?>
