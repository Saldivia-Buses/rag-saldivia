<?php
include ("./autoload.php");

include ("./sessionCheck.php");
//include ('../funciones/config.php');
include ('../funciones/utiles.php');

$DIR ='../atmail6/webmail/';

$db = $_SESSION["db"];
$database = Cache::getCache('datosbase'.$db);
if ($database === false){
    if ($xmlpath == '')  $xmlpath = '../database/';
    $config = new config('config.xml', $xmlpath, $db);
    $database = $config->bases[$db];
    Cache::setCache('datosbase'.$db, $database);
}



$imapServer= $database->imapServer;


$emailUser  = $_SESSION['emailUser'];
$emailPass  = $_SESSION['emailPass'];
$host= ($_SERVER['HTTPS'] =='on')?'https://':'http://';
$host.=$_SERVER['HTTP_HOST'];
//<div style="margin:20%;text-align:center; font-size:12px;font-weight:700;border:1px solid orange;background-color:#ffee7f;padding:10px;"><div id="throbber" /></div> Webmail</div>

?>
 <html>
 <head>
    <link rel="stylesheet" type="text/css" href="../funciones/concat.php?type=css" />
 
 </head>
 <body>
<div class="esperareloj"><b> Webmail</b><div id="throbber" /></div>
<form name="loginForm" target="_self" method="POST" action="../atmail6/webmail/index.php/mail/auth/processlogin">
<input type="hidden" name="emailName" value="<?php echo $emailUser; ?>" />
<input type="hidden" name="emailDomain" value="<?php echo $imapServer; ?>">
<input type="hidden" name="password" value="<?php echo $emailPass; ?>">
<input type="hidden" name="requestedServer" value="<?php echo $imapServer; ?>">
<input type="submit" value="GO" style="display:none;">
</form>

<script type="text/javascript">
    document.loginForm.submit();
</script>
</body>