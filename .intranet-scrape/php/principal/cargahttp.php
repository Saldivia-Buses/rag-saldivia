<?php
/*
 * Created on 23/10/2006
 * Luis M. Melgratti
 */
//include ("../includes/encab.php");
include ("./sessionCheck.php");

$url = $_GET['url'];
$newWin = $_GET['target'];
$get= parseGet($_GET);
//$newWin = false;
function parseGet($var, $nom=''){
        $get = '';
        global $newWin;
	foreach($var as $par => $val){
		if($par != 'url'){
			if (is_array($val)){
				$get.= parseGet($val, $par);
			}
			else {
				if ($nom != ''){
					$par = $nom.'['.$par.']';
				}
				$get.='&'.$par.'='.urlencode($val);
			}
		}
                if ($par == 'target' && $val == '_blank'){
                    $newWin = true;
                }
	}
	return $get;
}

if ($newWin == '_blank'){

?>
<body >
    <div class="Pagina">
        <div class="ingreso">
            <div class="error">
                    <?php
        	    echo "<a href=\"http://$url\" target=\"_blank\">$url</a>";
                    echo '<script language="JavaScript">';
                    echo "Histrix.loadExternalXML('gmail', 'http://$url');";
                    echo "</script>";
                ?>
            </div>
        </div>
    </div>
    </body>
<?php

}
else {
	$_SESSION['DAT'] = uniqid('DAT');
	$get .='&DAT='.$_SESSION['DAT'];

?>
<body >
<div class="consulta" style="width:100%">
<iframe class="contewin" style="width:100%;height:100%;" src="<?php echo $url.'?'.$get; ?>"></iframe>
</div>
</body>
<?php
}
?>
