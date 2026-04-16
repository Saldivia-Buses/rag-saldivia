<?php
include ("./sessionCheck.php");
$RoundCubeDIR ='../roundcube/';

$emailUser= $_SESSION['emailUser'];
$emailPass= $_SESSION['emailPass'];

?>
<div style="margin:20%;text-align:center; font-size:12px;font-weight:700;border:1px solid orange;background-color:#ffee7f;padding:10px;">
    WEBMAIL
</div>
<form name="f"
      action="<?php echo $RoundCubeDIR; ?>?_task=mail"
      method="post"
      style="display:none";>
      <input  name="_user" value="<?php echo $emailUser; ?>">
    <input  name="_pass" type="password" value="<?php echo $emailPass; ?>">
    <input name="_action" value="login" type="hidden" />
    <input name="_host" value="myhost" type="hidden">
    <input type="submit">
</form>
<script type="text/javascript">
    document.f.submit();
</script>
