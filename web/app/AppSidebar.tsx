"use client"

import {
    Sidebar,
    SidebarContent, SidebarFooter,
    SidebarGroup,
    SidebarGroupContent, SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu, SidebarMenuAction, SidebarMenuButton,
    SidebarMenuItem, SidebarMenuSub, SidebarMenuSubButton, SidebarMenuSubItem
} from "@/components/ui/sidebar";

import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "@/components/ui/collapsible"

import React from "react";
import {Users, LayoutDashboard, Boxes, KeySquare, Bolt, Book, ChevronRight} from "lucide-react";
import Image from "next/image";

const items = [
    {
        title: "Панель управления",
        url: "/dashboard",
        icon: LayoutDashboard,
    },
    {
        title: "Процессы",
        url: "/process",
        icon: Boxes,
        items:[
            {
                title: "Email",
                url: "/process/email",
            },
            {
                title: "Telegram",
                url: "/process/telegram",
            },
        ]
    },
    {
        title: "Пользователи",
        url: "/user",
        icon: Users,
    },
    {
        title: "Токены",
        url: "/token",
        icon: KeySquare,
    },
    {
        title: "Настройки",
        url: "#",
        icon: Bolt,
    },
    {
        title: "Документация",
        url: "#",
        icon: Book,
    },

]

export default function AppSidebar () : React.ReactNode {

    function logoClick (event: React.MouseEvent<HTMLButtonElement>): void {
        console.log(event)
    }

    return (
        <Sidebar>
            <SidebarHeader >
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton size="lg" asChild onClick={(e) => logoClick(e)}>
                            <a className={'flex flex-row items-center gap-2'} href={'/'}>
                                <Image width={32} height={32} src={`/logo.svg`} alt={`Кактус`}/> <span className={'text-2xl h-fit flex content-center'}>Cactus</span>
                            </a>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>

            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup >
                    <SidebarGroupLabel>Приложение</SidebarGroupLabel>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {items.map((item) => (
                                <Collapsible key={item.title} asChild>
                                    <SidebarMenuItem key={item.title}>
                                        <SidebarMenuButton   asChild>
                                            <a href={item.url}>
                                                <item.icon size={32}/>
                                                <span>{item.title}</span>
                                            </a>
                                        </SidebarMenuButton>
                                        {item.items?.length ? (
                                            <>
                                                <CollapsibleTrigger asChild>
                                                    <SidebarMenuAction className="data-[state=open]:rotate-90">
                                                        <ChevronRight />
                                                        <span className="sr-only">Toggle</span>
                                                    </SidebarMenuAction>
                                                </CollapsibleTrigger>
                                                <CollapsibleContent>
                                                    <SidebarMenuSub>
                                                        {item.items?.map((subItem) => (
                                                            <SidebarMenuSubItem key={subItem.title}>
                                                                <SidebarMenuSubButton asChild>
                                                                    <a href={subItem.url}>
                                                                        <span>{subItem.title}</span>
                                                                    </a>
                                                                </SidebarMenuSubButton>
                                                            </SidebarMenuSubItem>
                                                        ))}
                                                    </SidebarMenuSub>
                                                </CollapsibleContent>
                                            </>
                                        ) : null}
                                    </SidebarMenuItem>
                                </Collapsible>
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
                <SidebarGroup />
            </SidebarContent>
            <SidebarFooter></SidebarFooter>
        </Sidebar>
    )
}