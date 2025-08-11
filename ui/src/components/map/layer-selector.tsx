import { useWorld } from "@/hooks/use-world"
import { Layers } from "lucide-react"
import React from "react"

import { Button } from "../ui/button"
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from "../ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "../ui/popover"

export const LayerSelector: React.FC = () => {
    const { world, selectedDepth, selectDepth } = useWorld()
    const [open, setOpen] = React.useState(false)

    if (world.layers.length === 1) {
        return null
    }

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button variant="ghost" size="icon" className="h-6 text-xs px-6">
                    <Layers size={4} />
                    {selectedDepth}
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-0">
                <Command>
                    <CommandInput placeholder="Search layers..." className="h-9" />
                    <CommandList>
                        <CommandEmpty>No layers.</CommandEmpty>
                        <CommandGroup>
                            {world.layers.map((l) => (
                                <CommandItem
                                    key={l.depth}
                                    value={`${l.depth}`}
                                    onSelect={(v) => {
                                        selectDepth(parseInt(v))
                                        setOpen(false)
                                    }}
                                >
                                    {l.name || l.depth}
                                </CommandItem>
                            ))}
                        </CommandGroup>
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    )
}
