<?php



$options['all'] = 1;

if (!isset($jsmenu)) $jsmenu = '';

$sideMenu = new Histrix_Menu('menubar', 'v');
$sideMenu->build("_utilbar", 'myMenu2', $perfil, $jsmenu, 'menu2', $options);

$sideMenu->render();
 
?>

<div id="utilbarStatus2Div">
	<!-- <img  class="utilbarStatus" title="<?php echo $i18n['toggleUtilBar'];?>"  height="18px" width="18px" status="off" src="../img/atras.png" align="middle"> -->
</div>


<div style="position:absolute; bottom:0px;">
<div id="hipocampo"></div>
</div>