<?php
include('../webmail/integr.php');
include ("./sessionCheck.php");

$email      = $_SESSION['email'];
$login      = $_SESSION['emailUser'];
$password   = $_SESSION['emailPass'];
$imap       = $_SESSION['imapServer'];
$smtp       = $_SESSION['smtpServer'];


$Integr = new CIntegration('../webmail/');
if (!$Integr->UserLoginByEmail( $email, $login, START_PAGE_IS_MAILBOX, $password)){
    // Failed to log in, will try creating such account
    
    $acct = new Account();

    $acct->Email = $email;
    $acct->MailIncLogin = $login;
    $acct->MailIncHost = $imap;
    $acct->MailIncPassword = $password;
    $acct->MailIncPort = ($imap == 'imap.gmail.com')? 993 : 143;
    $acct->MailOutLogin = $login;
    $acct->MailOutHost = $smtp;
    $acct->MailOutPassword = $password;
    $acct->MailOutPort = 25;
    $acct->MailProtocol = MAILPROTOCOL_IMAP4;
    $acct->MailOutAuthentication = true;

    //if ($Integr->CreateUserFromAccount($acct))
    $User = new User();
    $User->CreateAccount($acct);
    if ($User->CreateUser($acct))
    {
        // nasty fix to create accounts
        $sql = 'UPDATE alogicmail.awm_account SET
                                def_acct = 1,id_user = id_acct;';

        consulta($sql);

        if (!$Integr->UserLoginByEmail($email, $login, START_PAGE_IS_MAILBOX, $password))
        {
            echo 'Error: the account has been created, but failed to log in though. Reason: ' . $Integr->GetErrorString();
        }
        else
        {
            echo 'Error: failed to log into the account. Reason: ' . $Integr->GetErrorString();
        }
    }
    else
    {
        echo 'Error: failed to create the account. Reason: ' . $Integr->GetErrorString();
    }
}

?>
 