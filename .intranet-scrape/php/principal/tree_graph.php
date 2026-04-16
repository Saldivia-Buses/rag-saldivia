<?php

/**
 *	phpTreeGraph
 *	Linux filesystem hierarchy demo
  * 	@author Mathias Herrmann
**/

//include GD rendering class
include_once ("./autoload.php");
require_once('../lib/phpTreeGraph/GDRenderer.php');
//include ("../classes/ContDatos.php");

include ("../funciones/utiles.php");
include ("./sessionCheck.php");

$xmlData = Histrix_XmlReader::unserializeContainer(null, $_REQUEST['instance']);
$Datatree = $xmlData->ARBOL;

//print_r($Datatree);
//create new GD renderer, optinal parameters: LevelSeparation,  SiblingSeparation, SubtreeSeparation, defaultNodeWidth, defaultNodeHeight
$objTree = new GDRenderer(30, 10, 30, 50, 40);
$i = 0;
$fontSize = 8;

function browseTree($Datatree, $parent){
	global $i;
	global $objTree;
	global $xmlData;

	global $fontSize;;

	if ($Datatree->nodos)
	foreach ($Datatree->nodos as $Nivel => $nodo){
		$arrayValues = $nodo->data;
		$text= '';
		$color = null;
		if ($arrayValues)
		foreach($arrayValues as $nval => $value){
			$ObjCampo = $xmlData->getCampo($nval);

			if ($ObjCampo->color == 'true' && $value != '' && $ObjCampo->textColor == 'true'){
				$rgb = hex_to_rgb($value);
				$color = array($rgb['red'], $rgb['green'], $rgb['blue']);
			}
			if ($ObjCampo->color == 'true' && $value != '' && $ObjCampo->titleColor == 'true'){
				$rgb = hex_to_rgb($value);
				$titleColor = array($rgb['red'], $rgb['green'], $rgb['blue']);
			}

			if ($titleColor == ''){
				if ($color){
					$r=$color[0] - 25;
					$g=$color[1] - 25;
					$b=$color[2] - 25;
					$titleColor = array($r, $g, $b);
				}
			}
			if ($ObjCampo->nodeTitle=="true"){
				$title = $value;

			}
			if ($ObjCampo->nodeText=="true"){
				$text .= $value;
			}
		}

		$i++;
		//echo($i.'_'.$parent.'_'.$text );
		//echo '<br>';

		$maxsize= max(strlen($text), strlen($title)) * $fontSize;

		$objTree->add($i,$parent,$title, $text , $maxsize, 40, null, $color, $titleColor);

		if ($nodo)
			browseTree($nodo, $i);
	}

}

browseTree($Datatree, 0);

//add nodes to the tree, parameters: id, parentid optional title, text, width, height, image(path)
//$objTree->setNodeLinks(GDRenderer::LINK_BEZIER);

$objTree->setBGColor(array(255, 255, 255));
$objTree->setNodeTitleColor(array(0, 128, 255));
$objTree->setNodeMessageColor(array(0, 192, 255));
$objTree->setLinkColor(array(0, 64, 128));
//$objTree->setNodeLinks(GDRenderer::LINK_BEZIER);
$objTree->setNodeBorder(array(0, 0, 0), 1);
$objTree->setFTFont('/usr/share/fonts/truetype/msttcorefonts/arial.ttf', $fontSize, 0, GDRenderer::CENTER);

$objTree->stream();
?>