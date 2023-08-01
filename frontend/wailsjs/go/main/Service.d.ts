// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {core} from '../models';
import {main} from '../models';

export function Connect(arg1:core.CONTROLLER_SESSION_REQUEST):Promise<main.ReturnObject>;

export function Disconnect():Promise<main.ReturnObject>;

export function ForwardToController(arg1:core.FORWARD_REQUEST):Promise<main.ReturnObject>;

export function ForwardToRouter(arg1:core.FORWARD_REQUEST):Promise<main.ReturnObject>;

export function GetLoadingLogs(arg1:string):Promise<main.ReturnObject>;

export function GetLogs(arg1:number):Promise<main.ReturnObject>;

export function GetQRCode(arg1:core.TWO_FACTOR_CONFIRM):Promise<main.ReturnObject>;

export function GetRoutersAndAccessPoints():Promise<main.ReturnObject>;

export function GetState():Promise<main.ReturnObject>;

export function OpenFileDialogForRouterFile(arg1:boolean):Promise<string>;

export function ResetEverything():Promise<main.ReturnObject>;

export function SetConfig(arg1:core.CONFIG_FORM):Promise<main.ReturnObject>;

export function Switch(arg1:core.CONTROLLER_SESSION_REQUEST):Promise<main.ReturnObject>;

export function SwitchRouter(arg1:string):Promise<main.ReturnObject>;