// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {core} from '../models';
import {main} from '../models';

export function BlockedDomainLogging(arg1:boolean):Promise<void>;

export function Connect(arg1:core.CONTROLLER_SESSION_REQUEST):Promise<main.ReturnObject>;

export function DisableAllBlocklists():Promise<void>;

export function DisableBlocklist(arg1:string):Promise<void>;

export function DisableDNSWhitelist():Promise<void>;

export function Disconnect():Promise<main.ReturnObject>;

export function EnableAllBlocklists():Promise<void>;

export function EnableBlocklist(arg1:string):Promise<void>;

export function EnableDNSWhitelist():Promise<void>;

export function ForwardToController(arg1:core.FORWARD_REQUEST):Promise<main.ReturnObject>;

export function ForwardToRouter(arg1:core.FORWARD_REQUEST):Promise<main.ReturnObject>;

export function GetLoadingLogs(arg1:string):Promise<main.ReturnObject>;

export function GetLogs(arg1:number):Promise<main.ReturnObject>;

export function GetQRCode(arg1:core.TWO_FACTOR_CONFIRM):Promise<main.ReturnObject>;

export function GetRoutersAndAccessPoints(arg1:core.FORWARD_REQUEST):Promise<main.ReturnObject>;

export function GetState():Promise<main.ReturnObject>;

export function LoadRoutersUnAuthenticated():Promise<main.ReturnObject>;

export function OpenFileDialogForRouterFile(arg1:boolean):Promise<string>;

export function RebuildDomainBlocklist():Promise<void>;

export function ResetEverything():Promise<main.ReturnObject>;

export function SetConfig(arg1:core.CONFIG_FORM):Promise<main.ReturnObject>;

export function StartDNSCapture():Promise<void>;

export function StopDNSCapture():Promise<string>;

export function Switch(arg1:core.CONTROLLER_SESSION_REQUEST):Promise<main.ReturnObject>;

export function SwitchRouter(arg1:string):Promise<main.ReturnObject>;
