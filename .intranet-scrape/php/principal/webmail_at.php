<?php
include ("./autoload.php");

include ("./sessionCheck.php");
//include ('../funciones/config.php');
include ('../funciones/utiles.php');

$DIR ='../atmail/';

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
$host= (isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] =='on')?'https://':'http://';
$host.=$_SERVER['HTTP_HOST'];
?>

<div style="margin:20%;text-align:center; font-size:12px;font-weight:700;border:1px solid orange;background-color:#ffee7f;padding:10px;">WEBMAIL</div>
 <html>
 <head>
<script language="javascript">

function LoginForm()    {

            var theForm = document.loginPage;

  if (theForm.username.value == "") 
  {
    alert('' + "You must enter your login name to enter Webmail on thisd server" + '');
    theForm.username.focus();
    return (false);
  }

  if (theForm.password.value == "")
  {
    alert('' + "Please enter your password" + '');
    theForm.username.focus();
    return (false);
  }

  if (!theForm.username.value || !theForm.pop3host.value || !theForm.password.value || !theForm.MailServer.value) return false;

    atmailroot = '<?php echo $host; ?>/atmail/';
       WebMailLoginReq = createXMLHttpRequest();
     //   WebMailLoginReq.onreadystatechange = WebMailLoginReqChange;
	WebMailLoginReq.onreadystatechange = function() {
	        if (WebMailLoginReq.readyState == 4) {
	            WebMailLoginReqChange(WebMailLoginReq.responseText);
	        }
	    }
var POSTString = "ajax=1&username=" + encodeURIComponent(theForm.username.value) + "&password=" + encodeURIComponent(theForm.password.value) + "&MailServer=" + encodeURIComponent(theForm.MailServer.value) + "&pop3host=" + encodeURIComponent(theForm.pop3host.value) + "&MailType=" + encodeURIComponent(theForm.MailType.value) + "&Language=" + "&LoginType=ajax";
   
    WebMailLoginReq.open("POST", atmailroot + "index.php", true);
    WebMailLoginReq.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
    WebMailLoginReq.send(POSTString);
   
}

function createXMLHttpRequest() {
    try { return new ActiveXObject("Msxml2.XMLHTTP"); } catch (e) {}
    try { return new ActiveXObject("Microsoft.XMLHTTP"); } catch (e) {}
    try { return new XMLHttpRequest(); } catch(e) {}
    alert("XMLHttpRequest not supported");
    return null;
}

function WebMailLoginReqChange() { 
                   location.href = atmailroot+'parse.php?file=html/LANG/simple/showmail_interface.html&ajax=1&func=Inbox&&To=';
}
</script> 
 </head>
 <body>
<form action="<?php echo $DIR; ?>atmail.php" method="post" target="_self" name="loginPage" _style="display:none;" />
	<input name="username" type="hidden" id="username" value="<?php echo $emailUser; ?>" />
	<input name="MailServer" type="hidden" class="logininput" id="MailServer" value="<?php echo $imapServer; ?>" />
	<input name="password" type="hidden" class="logininput" id="password" value="<?php echo $emailPass; ?>" />
    <input name="pop3host" type="hidden"  value="<?php echo $imapServer; ?>" />
    <input name="LoginType" type="hidden"  value="xp" />
    <input name="NewWindow" type="hidden"  value="1" />

    <input name="Language" type="hidden"  value="espanol" />
    <input name="MailType" type="hidden"  value="imap" />

    <input name='RememberMe'  type="hidden" value='1' />
	<input type="submit" name="Submit" value="Acceder" class="loginsubmit" />
</form>
				
<script type="text/javascript">
     document.loginPage.submit();
// LoginForm();
</script>
</body>
					