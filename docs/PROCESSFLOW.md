---
layout: default
title: Process Flow
nav_order: 2
permalink: /processflow
---

# Process Flow

The flowchar represents the internal process of the `repver` tool. It illustrates how the tool operates and the sequence of operations it performs to achieve its functionality.

## Initilization Phase

```mermaid
flowchart TD
    %% Initial configuration loading and parameter verification
    Start((Start)) --> DConfigExists{.repver exist?}
    DConfigExists -- Yes --> PLoadConfig[Load .repver]
    DConfigExists -- No --> ENoConfig[Error 100<br>.repver file not found]
    ENoConfig --> EndNoConfig((End))
    
    PLoadConfig --> DLoadSuccess{Load Success?}
    DLoadSuccess -- Yes --> PValidateConfig[Validate .repver]
    DLoadSuccess -- No --> ELoadFailed[Error 101<br>.repver failed to load]
    ELoadFailed --> EndLoadFailed((End))
    
    PValidateConfig --> DValidateSuccess{Validation Successful?}
    DValidateSuccess -- No --> EValidateFailed[Error 102<br>.repver validation failed]
    EValidateFailed --> EndValidateFailed((End))
    DValidateSuccess -- Yes --> PCommandArgs[Enumerate possible command line arguments from .repver]
    
    PCommandArgs --> PParseFlags[Parse command line arguments]
    PParseFlags --> DCommandSpecified{Command specified?}
    DCommandSpecified -- Yes --> PGetCommand[Retrieve command configuration]
    DCommandSpecified -- No --> ENoCommand[Error 103<br>No command specified]
    ENoCommand --> EndNoCommand((End))
    
    PGetCommand --> DCommandFound{Command found?}
    DCommandFound -- Yes --> PVerifyParams[Identify required arguments for command]
    DCommandFound -- No --> ECommandNotFound[Error 104<br>Command not found]
    ECommandNotFound --> EndCommandNotFound((End))
    
    PVerifyParams --> DParamsProvided{All params provided?}
    DParamsProvided -- No --> EMissingParams[Error 105<br>Missing required parameters]
    EMissingParams --> EndMissingParams((End))
    DParamsProvided -- Yes --> DGitOptionsProvided{Git options provided?}

    DGitOptionsProvided -- Yes --> DInGitRepo{In git repo?}
    DGitOptionsProvided -- No --> ExecPhase[Execution Phase]
    
    DInGitRepo -- No --> ENoGitRepo[Error 106<br>Not in git repository]
    ENoGitRepo --> EndNoGitRepo((End))
    DInGitRepo -- Yes --> ExecPhase
    
    %% Style definitions
    classDef startStyle fill:#d9f99d;
    classDef endStyle fill:#fecaca;
    classDef processStyle fill:#bfdbfe;
    classDef decisionStyle fill:#ddd6fe;
    
    %% Apply styles
    class Start startStyle;
    class EndNoConfig,EndLoadFailed,EndValidateFailed,EndNoCommand,EndCommandNotFound,EndMissingParams,EndNoGitRepo endStyle;
    class PLoadConfig,PValidateConfig,PCommandArgs,PParseFlags,PGetCommand,PVerifyParams,ExecPhase processStyle;
    class DConfigExists,DLoadSuccess,DValidateSuccess,DCommandSpecified,DCommandFound,DParamsProvided,DGitOptionsProvided,DInGitRepo decisionStyle;
```
